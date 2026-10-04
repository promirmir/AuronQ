package auronq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeaderOnlyValidationUsesFullNodeConsensusRules(t *testing.T) {
	easyPowForTest(t)
	c, _, miner := consensusTestChain(t, 3)

	history := []BlockHeader{c.network.Genesis.Header}
	for i := 0; i < 3; i++ {
		tpl, err := c.BuildTemplate(miner)
		if err != nil {
			t.Fatal(err)
		}
		mined := mineTemplateForInvariant(t, tpl)
		if err := ValidateHeaderEnvelope(mined.Header, history, time.Now().Unix()); err != nil {
			t.Fatalf("header %d rejected by light validator: %v", mined.Header.Height, err)
		}
		if err := c.AddBlock(&mined); err != nil {
			t.Fatalf("same block rejected by full node: %v", err)
		}
		history = append(history, mined.Header)
	}

	bad := history[len(history)-1]
	bad.Height++
	bad.PrevHash = ZeroHash()
	if err := ValidateHeaderEnvelope(bad, history, time.Now().Unix()); err == nil {
		t.Fatal("header with wrong prev hash was accepted")
	}
}

func TestP2PHeadersReturnsBoundedCanonicalHeaders(t *testing.T) {
	easyPowForTest(t)
	c, _, miner := consensusTestChain(t, 3)
	for i := 0; i < 2; i++ {
		tpl, err := c.BuildTemplate(miner)
		if err != nil {
			t.Fatal(err)
		}
		b := mineTemplateForInvariant(t, tpl)
		if err := c.AddBlock(&b); err != nil {
			t.Fatal(err)
		}
	}

	n := NewNode(c, NodeConfig{})
	req := httptest.NewRequest(http.MethodGet, "/p2p/headers?start=0&limit=3", nil)
	req.RemoteAddr = "8.8.8.8:45000"
	rec := httptest.NewRecorder()
	n.handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Start   uint64        `json:"start"`
		Headers []BlockHeader `json:"headers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Start != 0 || len(body.Headers) != 3 {
		t.Fatalf("unexpected response: start=%d headers=%d", body.Start, len(body.Headers))
	}
	for i, h := range body.Headers {
		if h.Height != uint64(i) {
			t.Fatalf("header[%d].height=%d", i, h.Height)
		}
	}
}
