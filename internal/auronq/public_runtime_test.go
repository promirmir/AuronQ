package auronq

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSetPublicAdvertiseDoesNotDeadlock(t *testing.T) {
	n := &Node{
		cfg:       NodeConfig{},
		announced: map[string]struct{}{"http://1.1.1.1:18444": {}},
	}
	done := make(chan bool, 1)
	go func() {
		done <- n.SetPublicAdvertise("http://8.8.8.8:18444")
	}()
	select {
	case ok := <-done:
		if !ok {
			t.Fatal("public advertise was rejected")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("SetPublicAdvertise deadlocked")
	}
	if got := n.PublicAdvertise(); got != "http://8.8.8.8:18444" {
		t.Fatalf("PublicAdvertise=%q", got)
	}
	if len(n.announced) != 0 {
		t.Fatal("announce cache was not cleared")
	}
}

func TestPermanentPortMappingCloseDeletesOnce(t *testing.T) {
	var deletes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(r.Header.Get("SOAPAction"), "DeletePortMapping") &&
			strings.Contains(string(body), "<NewExternalPort>18444</NewExternalPort>") {
			deletes.Add(1)
		}
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<ok/>"))
	}))
	defer srv.Close()

	m := &PortMapping{
		Advertise:   "http://8.8.8.8:18444",
		client:      srv.Client(),
		controlURL:  srv.URL,
		serviceType: "urn:schemas-upnp-org:service:WANIPConnection:1",
		port:        18444,
		lease:       0, // permanent mapping: no renewal goroutine
	}
	m.Close()
	m.Close()
	if got := deletes.Load(); got != 1 {
		t.Fatalf("DeletePortMapping calls=%d want=1", got)
	}
}
