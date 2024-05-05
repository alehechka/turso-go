package turso

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type SubscriptionClient client

type Subscription struct {
	Plan     string `json:"plan"`
	Timeline string `json:"timeline"`
	Overages bool   `json:"overages"`
}

func (c *SubscriptionClient) Get(ctx context.Context) (Subscription, error) {
	prefix := "/v1"
	if c.client.Org != "" {
		prefix = "/v1/organizations/" + c.client.Org
	}

	r, err := c.client.Get(ctx, prefix+"/subscription", nil)
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to get organization plan: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return Subscription{}, fmt.Errorf("failed to get organization plan with status %s: %v", r.Status, parseResponseError(r))
	}

	resp, err := unmarshal[struct{ Subscription Subscription }](r)
	return resp.Subscription, err
}

var ErrPaymentRequired = errors.New("payment required")

func (c *SubscriptionClient) Update(ctx context.Context, plan, timeline string, overages *bool) error {
	prefix := "/v1"
	if c.client.Org != "" {
		prefix = "/v1/organizations/" + c.client.Org
	}

	body, err := marshal(struct {
		Plan     string `json:"plan"`
		Timeline string `json:"timeline,omitempty"`
		Overages *bool  `json:"overages,omitempty"`
	}{plan, timeline, overages})
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}

	r, err := c.client.Post(ctx, prefix+"/subscription", body)
	if err != nil {
		return fmt.Errorf("failed to set organization plan: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusPaymentRequired {
		return ErrPaymentRequired
	}

	if r.StatusCode != 200 {
		return fmt.Errorf("failed to set organization plan with status %s: %v", r.Status, parseResponseError(r))
	}

	return nil
}
