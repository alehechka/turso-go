package turso

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Instance struct {
	Uuid     string
	Name     string
	Type     string
	Region   string
	Hostname string
}

type InstancesClient client

type CreateInstanceLocationError struct {
	err string
}

func (e *CreateInstanceLocationError) Error() string {
	return e.err
}

func (c *InstancesClient) List(ctx context.Context, db string) ([]Instance, error) {
	res, err := c.client.Get(ctx, c.URL(db, ""), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances of %s: %s", db, err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return nil, c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response with status code %d", res.StatusCode)
	}

	type ListResponse struct{ Instances []Instance }
	resp, err := unmarshal[ListResponse](res)
	if err != nil {
		return nil, err
	}

	return resp.Instances, nil
}

func (c *InstancesClient) Delete(ctx context.Context, db, instance string) error {
	url := c.URL(db, "/"+instance)
	res, err := c.client.Delete(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to destroy instances %s of %s: %s", instance, db, err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode == http.StatusBadRequest {
		body, _ := unmarshal[struct{ Error string }](res)
		return errors.New(body.Error)
	}

	if res.StatusCode == http.StatusNotFound {
		body, _ := unmarshal[struct{ Error string }](res)
		return errors.New(body.Error)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("response with status code %d", res.StatusCode)
	}

	return nil
}

func (c *InstancesClient) Create(ctx context.Context, dbName, location string) (*Instance, error) {
	type Body struct {
		Location string
	}
	body, err := marshal(Body{location})
	if err != nil {
		return nil, fmt.Errorf("could not serialize request body: %w", err)
	}

	url := c.URL(dbName, "")
	res, err := c.client.Post(ctx, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new instances for %s: %s", dbName, err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return nil, c.client.notMemberErr()
	}

	if res.StatusCode >= http.StatusInternalServerError {
		return nil, &CreateInstanceLocationError{fmt.Sprintf("failed to create new instance: %s", res.Status)}
	}

	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	data, err := unmarshal[struct{ Instance Instance }](res)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return &data.Instance, nil
}

func (c *InstancesClient) Wait(ctx context.Context, db, instance string) error {
	url := c.URL(db, "/"+instance+"/wait")
	res, err := c.client.Get(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for instance %s to of %s be ready: %s", instance, db, err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode == http.StatusBadRequest {
		body, _ := unmarshal[struct{ Error string }](res)
		return errors.New(body.Error)
	}

	if res.StatusCode == http.StatusNotFound {
		body, _ := unmarshal[struct{ Error string }](res)
		return errors.New(body.Error)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("response with status code %d", res.StatusCode)
	}

	return nil
}

func (d *InstancesClient) URL(database, suffix string) string {
	prefix := "/v1"
	if d.client.Org != "" {
		prefix = "/v1/organizations/" + d.client.Org
	}
	return fmt.Sprintf("%s/databases/%s/instances%s", prefix, database, suffix)
}
