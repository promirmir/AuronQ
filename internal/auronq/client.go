package auronq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Base string
	HTTP *http.Client
}

func NewClient(base string) *Client {
	return &Client{Base: strings.TrimRight(base, "/"), HTTP: &http.Client{Timeout: 30 * time.Second}}
}
func (c *Client) get(path string, out any) error {
	r, err := c.HTTP.Get(c.Base + path)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return decodeHTTPError(r)
	}
	return json.NewDecoder(r.Body).Decode(out)
}
func (c *Client) post(path string, in, out any) error {
	b, _ := json.Marshal(in)
	r, err := c.HTTP.Post(c.Base+path, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return decodeHTTPError(r)
	}
	return json.NewDecoder(r.Body).Decode(out)
}
func decodeHTTPError(r *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(r.Body, 8192))
	var v map[string]any
	if json.Unmarshal(b, &v) == nil {
		if e, ok := v["error"].(string); ok {
			return fmt.Errorf("node: %s", e)
		}
	}
	return fmt.Errorf("node HTTP %s: %s", r.Status, strings.TrimSpace(string(b)))
}
func (c *Client) Status() (Status, error) {
	var x Status
	err := c.get("/v1/status", &x)
	return x, err
}
func (c *Client) Balance(addr string) (BalanceResponse, error) {
	var x BalanceResponse
	err := c.get("/v1/balance?address="+url.QueryEscape(addr), &x)
	return x, err
}
func (c *Client) UTXOs(addr string) ([]UTXORecord, error) {
	var x []UTXORecord
	err := c.get("/v1/utxos?address="+url.QueryEscape(addr), &x)
	return x, err
}
func (c *Client) Template(addr string) (Block, error) {
	var x Block
	err := c.get("/v1/template?address="+url.QueryEscape(addr), &x)
	return x, err
}
func (c *Client) SubmitTx(tx Transaction) (TxResponse, error) {
	var x TxResponse
	err := c.post("/v1/tx", tx, &x)
	return x, err
}
func (c *Client) SubmitBlock(b Block) (BlockResponse, error) {
	var x BlockResponse
	err := c.post("/v1/block", b, &x)
	return x, err
}