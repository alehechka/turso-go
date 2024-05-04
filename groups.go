package turso

import (
	"context"
	"fmt"
	"net/http"
)

type GroupsClient client

type Group struct {
	Name      string   `json:"name"`
	Locations []string `json:"locations"`
	Primary   string   `json:"primary"`
	Archived  bool     `json:"archived"`
	Version   string   `json:"version"`
}

func (g *GroupsClient) List(ctx context.Context) ([]Group, error) {
	res, err := g.client.Get(ctx, g.URL(""), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return nil, g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get database groups: received status code%w", parseResponseError(res))
	}

	type ListResponse struct {
		Groups []Group `json:"groups"`
	}
	resp, err := unmarshal[ListResponse](res)
	return resp.Groups, err
}

func (g *GroupsClient) Get(ctx context.Context, name string) (Group, error) {
	res, err := g.client.Get(ctx, g.URL("/"+name), nil)
	if err != nil {
		return Group{}, fmt.Errorf("failed to get group %s: %w", name, err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return Group{}, g.client.notMemberErr()
	}

	if res.StatusCode == http.StatusNotFound {
		return Group{}, fmt.Errorf("group %s was not found", name)
	}

	if res.StatusCode != http.StatusOK {
		return Group{}, fmt.Errorf("failed to get database group: received status code%w", parseResponseError(res))
	}

	type Response struct {
		Group Group `json:"group"`
	}
	resp, err := unmarshal[Response](res)
	return resp.Group, err
}

func (g *GroupsClient) Delete(ctx context.Context, group string) error {
	url := g.URL("/" + group)
	res, err := g.client.Delete(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete group: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("group %s not found", group)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete group: received status code%w", parseResponseError(res))
	}

	return nil
}

func (g *GroupsClient) Create(ctx context.Context, name, location, version string) error {
	type Body struct{ Name, Location, Version string }
	body, err := marshal(Body{name, location, version})
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}

	res, err := g.client.Post(ctx, g.URL(""), body)
	if err != nil {
		return fmt.Errorf("failed to create group: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		return fmt.Errorf("group name '%s' is not available", name)
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

func (g *GroupsClient) Unarchive(ctx context.Context, name string) error {
	res, err := g.client.Post(ctx, g.URL("/"+name+"/unarchive"), nil)
	if err != nil {
		return fmt.Errorf("failed to unarchive group: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

func (g *GroupsClient) AddLocation(ctx context.Context, name, location string) error {
	res, err := g.client.Post(ctx, g.URL("/"+name+"/locations/"+location), nil)
	if err != nil {
		return fmt.Errorf("failed to post group location request: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

func (g *GroupsClient) RemoveLocation(ctx context.Context, name, location string) error {
	res, err := g.client.Delete(ctx, g.URL("/"+name+"/locations/"+location), nil)
	if err != nil {
		return fmt.Errorf("failed to post group location request: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

func (g *GroupsClient) WaitLocation(ctx context.Context, name, location string) error {
	res, err := g.client.Get(ctx, g.URL("/"+name+"/locations/"+location+"/wait"), nil)
	if err != nil {
		return fmt.Errorf("failed to send wait location request: %s", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

type Entities struct {
	DBNames []string `json:"databases,omitempty"`
}

type PermissionsClaim struct {
	ReadAttach Entities `json:"read_attach,omitempty"`
}

type GroupTokenRequest struct {
	Permissions *PermissionsClaim `json:"permissions,omitempty"`
}

func (g *GroupsClient) Token(ctx context.Context, group string, expiration string, readOnly bool, permissions *PermissionsClaim) (string, error) {
	authorization := ""
	if readOnly {
		authorization = "&authorization=read-only"
	}
	url := g.URL(fmt.Sprintf("/%s/auth/tokens?expiration=%s%s", group, expiration, authorization))

	req := GroupTokenRequest{permissions}
	body, err := marshal(req)
	if err != nil {
		return "", fmt.Errorf("could not serialize request body: %w", err)
	}

	res, err := g.client.Post(ctx, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to get database token: %w", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return "", g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get database token: %w", parseResponseError(res))
	}

	type JwtResponse struct{ Jwt string }
	data, err := unmarshal[JwtResponse](res)
	if err != nil {
		return "", err
	}
	return data.Jwt, nil
}

func (g *GroupsClient) Rotate(ctx context.Context, group string) error {
	url := g.URL(fmt.Sprintf("/%s/auth/rotate", group))
	res, err := g.client.Post(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to rotate database keys: %w", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to rotate database keys: %w", parseResponseError(res))
	}

	return nil
}

func (g *GroupsClient) Update(ctx context.Context, group string, version, extensions string) error {
	type Body struct{ Version, Extensions string }
	body, err := marshal(Body{version, extensions})
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}

	url := g.URL(fmt.Sprintf("/%s/update", group))
	res, err := g.client.Post(ctx, url, body)
	if err != nil {
		return fmt.Errorf("failed to rotate database keys: %w", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update group: %w", parseResponseError(res))
	}

	return nil
}

func (g *GroupsClient) Transfer(ctx context.Context, group string, to string) error {
	type Body struct {
		Organization string `json:"organization"`
	}
	body, err := marshal(Body{to})
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}

	url := g.URL(fmt.Sprintf("/%s/transfer", group))
	res, err := g.client.Post(ctx, url, body)
	if err != nil {
		return fmt.Errorf("failed to transfer group: %w", err)
	}
	defer res.Body.Close()

	if g.client.isNotMemberErr(res.StatusCode) {
		return g.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		err := parseResponseError(res)
		return fmt.Errorf("failed to transfer group: %w", err)
	}

	return nil
}

func (g *GroupsClient) URL(suffix string) string {
	prefix := "/v1"
	if g.client.Org != "" {
		prefix = "/v1/organizations/" + g.client.Org
	}
	return prefix + "/groups" + suffix
}
