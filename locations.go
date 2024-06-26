package turso

import (
	"context"
	"fmt"
	"net/http"
)

type LocationsClient client

type LocationsResponse struct {
	Locations map[string]string
}

type Location struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type LocationResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Closest     []Location
}

func (c *LocationsClient) List(ctx context.Context) (map[string]string, error) {
	r, err := c.client.Get(ctx, "/v1/locations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request locations: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get locations: %w", parseResponseError(r))

	}

	data, err := unmarshal[LocationsResponse](r)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize locations response: %w", err)
	}

	return data.Locations, nil
}

func (c *LocationsClient) Get(ctx context.Context, location string) (LocationResponse, error) {
	r, err := c.client.Get(ctx, "/v1/locations/"+location, nil)
	if err != nil {
		return LocationResponse{}, fmt.Errorf("failed to request location %s: %w", location, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return LocationResponse{}, fmt.Errorf("failed to get location %s: %s", location, r.Status)
	}

	data, err := unmarshal[struct {
		Location LocationResponse `json:"location"`
	}](r)

	if err != nil {
		return LocationResponse{}, fmt.Errorf("failed to deserialize location response: %w", err)
	}

	return data.Location, nil
}

type ClosestLocationResponse struct {
	Server string
}

func (c *LocationsClient) Closest(ctx context.Context) (string, error) {
	r, err := c.client.Get(ctx, "https://region.turso.io", nil)
	if err != nil {
		return "", fmt.Errorf("failed to request closest: %s", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get closest location: %w", parseResponseError(r))

	}

	data, err := unmarshal[ClosestLocationResponse](r)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize locations response: %w", err)
	}

	return data.Server, nil
}
