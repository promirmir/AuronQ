package auronq

import (
	"net"
	"testing"
)

func TestParseSSDPLocation(t *testing.T) {
	packet := []byte("HTTP/1.1 200 OK\r\nCACHE-CONTROL: max-age=120\r\nLoCaTiOn: http://192.168.1.1:49000/rootDesc.xml\r\n\r\n")
	got := parseSSDPLocation(packet)
	if got != "http://192.168.1.1:49000/rootDesc.xml" {
		t.Fatalf("location=%q", got)
	}
}

func TestLocalRouterURLRejectsPublicAndAcceptsPrivate(t *testing.T) {
	if _, _, err := localRouterURL("http://192.168.1.1:1900/root.xml"); err != nil {
		t.Fatalf("private router rejected: %v", err)
	}
	if _, _, err := localRouterURL("http://8.8.8.8/root.xml"); err == nil {
		t.Fatal("public UPnP URL accepted")
	}
	if _, _, err := localRouterURL("http://router.example/root.xml"); err == nil {
		t.Fatal("hostname UPnP URL accepted")
	}
}

func TestFindWANServiceNested(t *testing.T) {
	want := upnpService{ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:1", ControlURL: "/ctl/IPConn"}
	root := upnpDevice{Devices: []upnpDevice{{Services: []upnpService{want}}}}
	got, ok := findWANService(root)
	if !ok || got != want {
		t.Fatalf("service=%+v ok=%v", got, ok)
	}
}

func TestXMLTextAndPublicExternalIPRules(t *testing.T) {
	body := []byte(`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><u:GetExternalIPAddressResponse xmlns:u="urn:test"><NewExternalIPAddress>203.0.113.50</NewExternalIPAddress></u:GetExternalIPAddressResponse></s:Body></s:Envelope>`)
	if got := xmlText(body, "NewExternalIPAddress"); got != "203.0.113.50" {
		t.Fatalf("external ip=%q", got)
	}
	if !isNonPublicIP(net.ParseIP("100.64.1.1")) {
		t.Fatal("CGNAT address considered public")
	}
}
