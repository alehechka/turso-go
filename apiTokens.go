package turso

import (
	"context"
	"fmt"
	"net/http"
)

type ApiToken struct {
	ID     string `json:"dbId"`
	Name   string
	Owner  uint
	PubKey []byte
}

type ApiTokensClient client

func (c *ApiTokensClient) List(ctx context.Context) ([]ApiToken, error) {
	res, err := c.client.Get(ctx, "/v1/auth/api-tokens", nil)
	if err != nil {
		return []ApiToken{}, fmt.Errorf("failed to get api tokens list: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get api tokens list: %s", res.Status)
	}

	type ListResponse struct {
		ApiTokens []ApiToken `json:"tokens"`
	}
	resp, err := unmarshal[ListResponse](res)
	return resp.ApiTokens, err
}

type CreateApiToken struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

func (c *ApiTokensClient) Create(ctx context.Context, name string) (CreateApiToken, error) {
	url := fmt.Sprintf("/v2/auth/api-tokens/%s", name)

	res, err := c.client.Post(ctx, url, nil)
	if err != nil {
		return CreateApiToken{}, fmt.Errorf("failed to create token: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return CreateApiToken{}, parseResponseError(res)
	}

	type CreateApiTokenResponse struct {
		ApiToken CreateApiToken `json:"token"`
	}

	data, err := unmarshal[CreateApiTokenResponse](res)
	if err != nil {
		return CreateApiToken{}, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return data.ApiToken, nil
}

func (c *ApiTokensClient) Revoke(ctx context.Context, name string) error {
	url := fmt.Sprintf("/v1/auth/api-tokens/%s", name)

	res, err := c.client.Delete(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to revoke API token: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}

	return nil
}
