package turso

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Database struct {
	ID            string `json:"dbId" mapstructure:"dbId"`
	Name          string
	Regions       []string
	PrimaryRegion string
	Hostname      string
	Version       string
	Group         string
	Sleeping      bool
}

type DatabasesClient client

func (c *DatabasesClient) List(ctx context.Context) ([]Database, error) {
	res, err := c.client.Get(ctx, c.URL(""), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get database listing: %s", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return nil, c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get database listing: %w", parseResponseError(res))
	}

	type ListResponse struct {
		Databases []Database `json:"databases"`
	}
	resp, err := unmarshal[ListResponse](res)
	return resp.Databases, err
}

func (c *DatabasesClient) Delete(ctx context.Context, database string) error {
	url := c.URL("/" + database)
	res, err := c.client.Delete(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete database: %s", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("database %s not found", database)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete database: %w", parseResponseError(res))
	}

	return nil
}

type CreateDatabaseResponse struct {
	Database Database
	Username string
}

type DBSeed struct {
	Type      string     `json:"type"`
	Name      string     `json:"value,omitempty"`
	URL       string     `json:"url,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type CreateDatabaseBody struct {
	Name       string  `json:"name"`
	Location   string  `json:"location"`
	Image      string  `json:"image,omitempty"`
	Extensions string  `json:"extensions,omitempty"`
	Group      string  `json:"group,omitempty"`
	Seed       *DBSeed `json:"seed,omitempty"`
	Schema     string  `json:"schema,omitempty"`
	IsSchema   bool    `json:"is_schema,omitempty"`
}

func (c *DatabasesClient) Create(ctx context.Context, name, location, image, extensions, group string, schema string, isSchema bool, seed *DBSeed) (*CreateDatabaseResponse, error) {
	params := CreateDatabaseBody{name, location, image, extensions, group, seed, schema, isSchema}

	body, err := marshal(params)
	if err != nil {
		return nil, fmt.Errorf("could not serialize request body: %w", err)
	}

	res, err := c.client.Post(ctx, c.URL(""), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %s", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return nil, c.client.notMemberErr()
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		return nil, fmt.Errorf("database name '%s' is not available", name)
	}

	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	data, err := unmarshal[*CreateDatabaseResponse](res)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	return data, nil
}

func (c *DatabasesClient) Seed(ctx context.Context, name string, dbFile *os.File) error {
	url := c.URL(fmt.Sprintf("/%s/seed", name))
	res, err := c.client.Upload(ctx, url, dbFile)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		return fmt.Errorf("database name '%s' is not available", name)
	}

	if res.StatusCode != http.StatusOK {
		return parseResponseError(res)
	}
	return nil
}

func (c *DatabasesClient) UploadDump(ctx context.Context, dbFile *os.File) (string, error) {
	url := c.URL("/dumps")
	res, err := c.client.Upload(ctx, url, dbFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload the dump file: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return "", c.client.notMemberErr()
	}
	if res.StatusCode != http.StatusOK {
		return "", parseResponseError(res)
	}
	type response struct {
		DumpURL string `json:"dump_url"`
	}
	data, err := unmarshal[response](res)
	if err != nil {
		return "", err
	}
	return data.DumpURL, nil
}

type DatabaseTokenRequest struct {
	Permissions *PermissionsClaim `json:"permissions,omitempty"`
}

func (c *DatabasesClient) Token(ctx context.Context, database string, expiration string, readOnly bool, permissions *PermissionsClaim) (string, error) {
	authorization := ""
	if readOnly {
		authorization = "&authorization=read-only"
	}
	url := c.URL(fmt.Sprintf("/%s/auth/tokens?expiration=%s%s", database, expiration, authorization))

	req := DatabaseTokenRequest{permissions}
	body, err := marshal(req)
	if err != nil {
		return "", fmt.Errorf("could not serialize request body: %w", err)
	}

	res, err := c.client.Post(ctx, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to get database token: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return "", c.client.notMemberErr()
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

func (c *DatabasesClient) Rotate(ctx context.Context, database string) error {
	url := c.URL(fmt.Sprintf("/%s/auth/rotate", database))
	res, err := c.client.Post(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to rotate database keys: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to rotate database keys: %w", parseResponseError(res))
	}

	return nil
}

func (c *DatabasesClient) Update(ctx context.Context, database string, group bool) error {
	url := c.URL(fmt.Sprintf("/%s/update", database))
	if group {
		url += "?group=true"
	}
	res, err := c.client.Post(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update database: %w", parseResponseError(res))
	}

	return nil
}

type Stats struct {
	TopQueries []struct {
		Query       string `json:"query"`
		RowsRead    int    `json:"rows_read"`
		RowsWritten int    `json:"rows_written"`
	} `json:"top_queries,omitempty"`
}

func (c *DatabasesClient) Stats(ctx context.Context, database string) (Stats, error) {
	url := c.URL(fmt.Sprintf("/%s/stats", database))
	res, err := c.client.Get(ctx, url, nil)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to update database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return Stats{}, c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return Stats{}, fmt.Errorf("failed to get stats for database: %w", parseResponseError(res))
	}

	return unmarshal[Stats](res)
}

type Body struct {
	Org string `json:"org"`
}

func (c *DatabasesClient) Transfer(ctx context.Context, database, org string) error {
	url := c.URL(fmt.Sprintf("/%s/transfer", database))
	body, err := json.Marshal(Body{Org: org})
	bodyReader := bytes.NewReader(body)
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}
	res, err := c.client.Post(ctx, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to transfer database")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to transfer %s database to org %s: %w", database, org, parseResponseError(res))
	}

	return nil
}

func (c *DatabasesClient) Wakeup(ctx context.Context, database string) error {
	url := c.URL(fmt.Sprintf("/%s/wakeup", database))
	res, err := c.client.Post(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to wakeup database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to wakeup database: %w", parseResponseError(res))
	}

	return nil
}

type Usage struct {
	RowsRead         uint64 `json:"rows_read,omitempty"`
	RowsWritten      uint64 `json:"rows_written,omitempty"`
	StorageBytesUsed uint64 `json:"storage_bytes,omitempty"`
	BytesSynced      uint64 `json:"bytes_synced,omitempty"`
}

type InstanceUsage struct {
	UUID  string `json:"uuid,omitempty"`
	Usage Usage  `json:"usage"`
}

type DbUsage struct {
	UUID      string          `json:"uuid,omitempty"`
	Instances []InstanceUsage `json:"instances"`
	Usage     Usage           `json:"usage"`
}

type DbUsageResponse struct {
	DbUsage DbUsage `json:"database"`
}

func (c *DatabasesClient) Usage(ctx context.Context, database string) (DbUsage, error) {
	url := c.URL(fmt.Sprintf("/%s/usage", database))

	res, err := c.client.Get(ctx, url, nil)
	if err != nil {
		return DbUsage{}, fmt.Errorf("failed to get database usage: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return DbUsage{}, fmt.Errorf("failed to get database usage: %w", parseResponseError(res))
	}

	body, err := unmarshal[DbUsageResponse](res)
	if err != nil {
		return DbUsage{}, err
	}
	return body.DbUsage, nil
}

func (c *DatabasesClient) URL(suffix string) string {
	prefix := "/v1"
	if c.client.Org != "" {
		prefix = "/v1/organizations/" + c.client.Org
	}
	return prefix + "/databases" + suffix
}

type DatabaseConfig struct {
	AllowAttach bool `json:"allow_attach"`
}

func (c *DatabasesClient) GetConfig(ctx context.Context, database string) (DatabaseConfig, error) {
	url := c.URL(fmt.Sprintf("/%s/configuration", database))
	res, err := c.client.Get(ctx, url, nil)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to get database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return DatabaseConfig{}, c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		err = parseResponseError(res)
		return DatabaseConfig{}, fmt.Errorf("failed to get config for database: %d %s", res.StatusCode, err)
	}

	return unmarshal[DatabaseConfig](res)
}

func (c *DatabasesClient) UpdateConfig(ctx context.Context, database string, config DatabaseConfig) error {
	url := c.URL(fmt.Sprintf("/%s/configuration", database))
	body, err := marshal(config)
	if err != nil {
		return fmt.Errorf("could not serialize request body: %w", err)
	}
	res, err := c.client.Patch(ctx, url, body)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}
	defer res.Body.Close()

	if c.client.isNotMemberErr(res.StatusCode) {
		return c.client.notMemberErr()
	}

	if res.StatusCode != http.StatusOK {
		err = parseResponseError(res)
		return fmt.Errorf("failed to update config for database: %d %s", res.StatusCode, err)
	}

	return nil
}
