package turso

import (
	"context"
	"fmt"
	"net/http"
)

type TokensClient client

func (c *TokensClient) Validate(ctx context.Context, token string) (int64, error) {
	r, err := c.client.Get(ctx, "/v1/auth/validate", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to request validation: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to validate token: %w", parseResponseError(r))
	}

	data, err := unmarshal[struct{ Exp int64 }](r)
	if err != nil {
		return 0, fmt.Errorf("failed to deserialize validate token response: %w", err)
	}

	return data.Exp, nil
}

func (c *TokensClient) Invalidate(ctx context.Context) (int64, error) {
	r, err := c.client.Post(ctx, "/v1/auth/invalidate", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to request invalidation: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to invalidate sessions: %w", parseResponseError(r))
	}

	data, err := unmarshal[struct{ ValidFrom int64 }](r)
	if err != nil {
		return 0, fmt.Errorf("failed to deserialize invalidate sessions response: %w", err)
	}

	return data.ValidFrom, nil
}
