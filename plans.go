package turso

import (
	"context"
	"fmt"
)

type PlansClient client

type Plan struct {
	Name   string `json:"name"`
	Price  string `json:"price"`
	Quotas struct {
		RowsRead    uint64 `json:"rowsRead"`
		RowsWritten uint64 `json:"rowsWritten"`
		Databases   uint64 `json:"databases"`
		BytesSynced uint64 `json:"bytesSynced"`
		Locations   uint64 `json:"locations"`
		Storage     uint64 `json:"storage"`
		Groups      uint64 `json:"groups"`
	}
}

func (c *PlansClient) List(ctx context.Context) ([]Plan, error) {
	r, err := c.client.Get(ctx, "/v1/plans", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan list: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("failed to list plans with status %s: %v", r.Status, parseResponseError(r))
	}

	resp, err := unmarshal[struct{ Plans []Plan }](r)
	return resp.Plans, err
}
