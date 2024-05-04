package turso

import (
	"context"
	"fmt"
	"net/http"
)

type InvoicesClient client

type Invoice struct {
	Number          string `json:"invoice_number"`
	Amount          string `json:"amount_due"`
	DueDate         string `json:"due_date"`
	PaidAt          string `json:"paid_at"`
	PaymentFailedAt string `json:"payment_failed_at"`
	InvoicePdf      string `json:"invoice_pdf"`
}

func (c *InvoicesClient) List(ctx context.Context) ([]Invoice, error) {
	res, err := c.client.Get(ctx, c.URL(""), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return nil, c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get invoices: received status code%w", parseResponseError(res))
	}

	type ListResponse struct {
		Invoices []Invoice `json:"invoices"`
	}
	resp, err := unmarshal[ListResponse](res)
	return resp.Invoices, err
}

func (c *InvoicesClient) URL(suffix string) string {
	prefix := "/v1"
	if c.client.Org != "" {
		prefix = "/v1/organizations/" + c.client.Org
	}
	return prefix + "/invoices" + suffix
}
