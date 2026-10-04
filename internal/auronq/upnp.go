package auronq

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ssdpAddress          = "239.255.255.250:1900"
	upnpDiscoveryWindow  = 2200 * time.Millisecond
	upnpHTTPTimeout      = 3 * time.Second
	upnpDescriptionLimit = 512 * 1024
	upnpSOAPLimit        = 256 * 1024
)

type upnpService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

type upnpDevice struct {
	Services []upnpService `xml:"serviceList>service"`
	Devices  []upnpDevice  `xml:"deviceList>device"`
}

type upnpRoot struct {
	Device upnpDevice `xml:"device"`
}

// PortMapping represents an automatically created router mapping.
// Close is idempotent and best-effort.
type PortMapping struct {
	Advertise   string
	client      *http.Client
	controlURL  string
	serviceType string
	port        uint16
	lease       uint32

	mu     sync.Mutex
	stop   chan struct{}
	done   chan struct{}
	closed bool
}

func (m *PortMapping) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	stop := m.stop
	done := m.done
	m.stop = nil
	m.done = nil
	if stop != nil {
		close(stop)
	}
	m.mu.Unlock()

	if done != nil {
		select {
		case <-done:
		case <-time.After(upnpHTTPTimeout + 500*time.Millisecond):
		}
	}

	// Permanent mappings (lease == 0) do not run a renewal goroutine, but they
	// still must be removed when the application/node shuts down.
	if m.client != nil && m.controlURL != "" && m.serviceType != "" && m.port != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), upnpHTTPTimeout)
		defer cancel()
		_ = deletePortMapping(ctx, m.client, m.controlURL, m.serviceType, m.port)
	}
}

func (m *PortMapping) startRenewal(localIP string) {
	if m == nil || m.lease == 0 {
		return
	}
	m.mu.Lock()
	if m.stop != nil {
		m.mu.Unlock()
		return
	}
	m.stop = make(chan struct{})
	m.done = make(chan struct{})
	stop := m.stop
	done := m.done
	lease := m.lease
	m.mu.Unlock()

	interval := time.Duration(lease) * time.Second / 2
	if interval < 5*time.Minute {
		interval = 5 * time.Minute
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), upnpHTTPTimeout)
				_ = addPortMapping(ctx, m.client, m.controlURL, m.serviceType, localIP, m.port, lease)
				cancel()
			}
		}
	}()
}

func parseSSDPLocation(packet []byte) string {
	for _, line := range strings.Split(strings.ReplaceAll(string(packet), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 9 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(strings.SplitN(line, ":", 2)[0]), "location") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func localRouterURL(raw string) (*url.URL, net.IP, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" {
		return nil, nil, errors.New("invalid UPnP URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil, errors.New("unsupported UPnP URL scheme")
	}
	host := strings.Trim(u.Hostname(), "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, nil, errors.New("UPnP URL must use a local literal IP")
	}
	if !(ip.IsPrivate() || ip.IsLinkLocalUnicast()) || ip.IsLoopback() || ip.IsUnspecified() {
		return nil, nil, errors.New("UPnP URL is not local")
	}
	return u, ip, nil
}

func findWANService(d upnpDevice) (upnpService, bool) {
	for _, s := range d.Services {
		t := strings.ToLower(strings.TrimSpace(s.ServiceType))
		if strings.Contains(t, ":service:wanipconnection:") || strings.Contains(t, ":service:wanpppconnection:") {
			return s, true
		}
	}
	for _, child := range d.Devices {
		if s, ok := findWANService(child); ok {
			return s, true
		}
	}
	return upnpService{}, false
}

func localIPForRouter(routerIP net.IP) (string, error) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: routerIP, Port: 9})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP == nil || addr.IP.IsUnspecified() {
		return "", errors.New("cannot determine local IP for router")
	}
	return addr.IP.String(), nil
}

func upnpHTTPClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = nil
	return &http.Client{
		Timeout:   upnpHTTPTimeout,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func fetchWANService(ctx context.Context, client *http.Client, location string) (string, string, string, error) {
	base, routerIP, err := localRouterURL(location)
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return "", "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("UPnP description HTTP %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, upnpDescriptionLimit+1))
	if err != nil || len(body) > upnpDescriptionLimit {
		return "", "", "", errors.New("UPnP description too large or unreadable")
	}
	var root upnpRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return "", "", "", err
	}
	svc, ok := findWANService(root.Device)
	if !ok || strings.TrimSpace(svc.ControlURL) == "" || strings.TrimSpace(svc.ServiceType) == "" {
		return "", "", "", errors.New("UPnP WAN connection service not found")
	}
	controlRef, err := url.Parse(strings.TrimSpace(svc.ControlURL))
	if err != nil {
		return "", "", "", err
	}
	control := base.ResolveReference(controlRef)
	checked, controlIP, err := localRouterURL(control.String())
	if err != nil || !controlIP.Equal(routerIP) {
		return "", "", "", errors.New("unsafe UPnP control URL")
	}
	localIP, err := localIPForRouter(routerIP)
	if err != nil {
		return "", "", "", err
	}
	return checked.String(), strings.TrimSpace(svc.ServiceType), localIP, nil
}

func xmlText(body []byte, localName string) string {
	dec := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != localName {
			continue
		}
		var text string
		if dec.DecodeElement(&text, &se) == nil {
			return strings.TrimSpace(text)
		}
		return ""
	}
}

func soapCall(ctx context.Context, client *http.Client, controlURL, serviceType, action, args string) ([]byte, error) {
	envelope := `<?xml version="1.0"?>` +
		`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">` +
		`<s:Body><u:` + action + ` xmlns:u="` + serviceType + `">` + args + `</u:` + action + `></s:Body></s:Envelope>`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, controlURL, strings.NewReader(envelope))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"`+serviceType+`#`+action+`"`)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, upnpSOAPLimit+1))
	if readErr != nil || len(body) > upnpSOAPLimit {
		return nil, errors.New("UPnP SOAP response too large or unreadable")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("UPnP %s HTTP %s", action, resp.Status)
	}
	return body, nil
}

func getExternalIPAddress(ctx context.Context, client *http.Client, controlURL, serviceType string) (string, error) {
	body, err := soapCall(ctx, client, controlURL, serviceType, "GetExternalIPAddress", "")
	if err != nil {
		return "", err
	}
	raw := xmlText(body, "NewExternalIPAddress")
	ip := net.ParseIP(strings.TrimSpace(raw))
	if isNonPublicIP(ip) {
		return "", errors.New("router does not expose a public WAN IP (NAT/CGNAT likely)")
	}
	return ip.String(), nil
}

func addPortMapping(ctx context.Context, client *http.Client, controlURL, serviceType, localIP string, port uint16, lease uint32) error {
	p := strconv.Itoa(int(port))
	args := "<NewRemoteHost></NewRemoteHost>" +
		"<NewExternalPort>" + p + "</NewExternalPort>" +
		"<NewProtocol>TCP</NewProtocol>" +
		"<NewInternalPort>" + p + "</NewInternalPort>" +
		"<NewInternalClient>" + localIP + "</NewInternalClient>" +
		"<NewEnabled>1</NewEnabled>" +
		"<NewPortMappingDescription>AuronQ Full Node</NewPortMappingDescription>" +
		"<NewLeaseDuration>" + strconv.FormatUint(uint64(lease), 10) + "</NewLeaseDuration>"
	_, err := soapCall(ctx, client, controlURL, serviceType, "AddPortMapping", args)
	return err
}

func deletePortMapping(ctx context.Context, client *http.Client, controlURL, serviceType string, port uint16) error {
	p := strconv.Itoa(int(port))
	args := "<NewRemoteHost></NewRemoteHost>" +
		"<NewExternalPort>" + p + "</NewExternalPort>" +
		"<NewProtocol>TCP</NewProtocol>"
	_, err := soapCall(ctx, client, controlURL, serviceType, "DeletePortMapping", args)
	return err
}

// TryUPnPPortMapping asks a local Internet Gateway Device to map TCP port to
// this machine. It returns only when the gateway also reports a globally
// routable WAN address. Remote AuronQ peers still callback-verify reachability
// before admitting the endpoint to public gossip.
func TryUPnPPortMapping(ctx context.Context, port uint16) (*PortMapping, error) {
	if port == 0 {
		return nil, errors.New("invalid port")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	deadline := time.Now().Add(upnpDiscoveryWindow)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	dst, _ := net.ResolveUDPAddr("udp4", ssdpAddress)
	searchTargets := []string{
		"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
		"urn:schemas-upnp-org:device:InternetGatewayDevice:2",
	}
	for _, st := range searchTargets {
		msg := "M-SEARCH * HTTP/1.1\r\n" +
			"HOST: " + ssdpAddress + "\r\n" +
			"MAN: \"ssdp:discover\"\r\n" +
			"MX: 1\r\n" +
			"ST: " + st + "\r\n\r\n"
		_, _ = conn.WriteToUDP([]byte(msg), dst)
	}

	client := upnpHTTPClient()
	seen := map[string]bool{}
	var lastErr error
	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ctx.Err()
		default:
		}
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, errors.New("UPnP/IGD not available on this network")
		}
		location := parseSSDPLocation(buf[:n])
		if location == "" || seen[location] {
			continue
		}
		seen[location] = true

		tryCtx, cancel := context.WithTimeout(ctx, upnpHTTPTimeout)
		controlURL, serviceType, localIP, err := fetchWANService(tryCtx, client, location)
		cancel()
		if err != nil {
			lastErr = err
			continue
		}

		lease := uint32(0)
		mapCtx, cancel := context.WithTimeout(ctx, upnpHTTPTimeout)
		err = addPortMapping(mapCtx, client, controlURL, serviceType, localIP, port, lease)
		cancel()
		if err != nil {
			lease = 3600
			mapCtx, cancel = context.WithTimeout(ctx, upnpHTTPTimeout)
			err = addPortMapping(mapCtx, client, controlURL, serviceType, localIP, port, lease)
			cancel()
		}
		if err != nil {
			lastErr = err
			continue
		}

		ipCtx, cancel := context.WithTimeout(ctx, upnpHTTPTimeout)
		externalIP, err := getExternalIPAddress(ipCtx, client, controlURL, serviceType)
		cancel()
		if err != nil {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), upnpHTTPTimeout)
			_ = deletePortMapping(cleanupCtx, client, controlURL, serviceType, port)
			cleanupCancel()
			lastErr = err
			continue
		}

		m := &PortMapping{
			Advertise:   "http://" + net.JoinHostPort(externalIP, strconv.Itoa(int(port))),
			client:      client,
			controlURL:  controlURL,
			serviceType: serviceType,
			port:        port,
			lease:       lease,
		}
		m.startRenewal(localIP)
		return m, nil
	}
}
