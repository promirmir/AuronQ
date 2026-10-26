package main

import (
	"context"
	"time"

	aq "auronq/internal/auronq"
)

// nodeRecoveryDelay caps reconnect backoff without returning to mining on
// stale node state. Thermal and backend errors are handled separately.
func nodeRecoveryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt >= 12 {
		return 60 * time.Second
	}
	d := time.Duration(attempt+1) * 5 * time.Second
	if d > 60*time.Second {
		return 60 * time.Second
	}
	return d
}

type minerNodeResolver struct {
	network   *aq.NetworkConfig
	preferred string
	store     string
	endpoint  string
	client    *aq.Client
	status    aq.Status
}

func (r *minerNodeResolver) connect(ctx context.Context, excludeCurrent bool) error {
	var exclude []string
	if excludeCurrent && r.endpoint != "" {
		exclude = []string{r.endpoint}
	}
	res, err := aq.ResolveVerifiedClient(ctx, aq.VerifiedClientConfig{
		ExpectedNetworkID: r.network.NetworkID(),
		Preferred:         r.preferred,
		Seeds:             r.network.SeedPeers,
		StorePath:         r.store,
		Exclude:           exclude,
		Timeout:           4 * time.Second,
	})
	if err != nil {
		return err
	}
	r.endpoint = res.Endpoint
	r.client = res.Client
	r.status = res.Status
	return nil
}
