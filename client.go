package turso

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
)

// Collection of all turso clients
type Client struct {
	baseUrl    string
	token      string
	Org        string
	version    string
	httpClient *http.Client

	// Single instance to be reused by all clients
	base *client

	Instances     *InstancesClient
	Databases     *DatabasesClient
	Feedback      *FeedbackClient
	Organizations *OrganizationsClient
	ApiTokens     *ApiTokensClient
	Locations     *LocationsClient
	Tokens        *TokensClient
	Users         *UsersClient
	Plans         *PlansClient
	Subscriptions *SubscriptionClient
	Billing       *BillingClient
	Groups        *GroupsClient
	Invoices      *InvoicesClient
}

// Client struct that will be aliases by all other clients
type client struct {
	client *Client
}

type ClientOption interface {
	apply(*Client)
}

const BaseURL = "https://api.turso.tech"

func New(token string, org string, options ...ClientOption) (*Client, error) {
	c := &Client{baseUrl: BaseURL, token: token, Org: org, httpClient: http.DefaultClient, version: getVersion()}

	for _, option := range options {
		option.apply(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	c.base = &client{c}
	c.Instances = (*InstancesClient)(c.base)
	c.Databases = (*DatabasesClient)(c.base)
	c.Feedback = (*FeedbackClient)(c.base)
	c.Organizations = (*OrganizationsClient)(c.base)
	c.ApiTokens = (*ApiTokensClient)(c.base)
	c.Locations = (*LocationsClient)(c.base)
	c.Tokens = (*TokensClient)(c.base)
	c.Users = (*UsersClient)(c.base)
	c.Plans = (*PlansClient)(c.base)
	c.Subscriptions = (*SubscriptionClient)(c.base)
	c.Billing = (*BillingClient)(c.base)
	c.Groups = (*GroupsClient)(c.base)
	c.Invoices = (*InvoicesClient)(c.base)
	return c, nil
}

func (c *Client) validate() error {
	if c.baseUrl == "" {
		return errors.New("no baseUrl set")
	}

	if c.token == "" {
		return errors.New("no API token set")
	}

	if c.httpClient == nil {
		return errors.New("no httpClient set")
	}

	return nil
}

func getVersion() (version string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	module := "github.com/alehechka/turso-go"
	for _, dep := range bi.Deps {
		if dep.Path == module {
			return dep.Version
		}
	}

	return
}

type withHTTPClient struct {
	httpClient *http.Client
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return &withHTTPClient{httpClient: httpClient}
}

func (o *withHTTPClient) apply(client *Client) {
	client.httpClient = o.httpClient
}

type withBaseUrl struct {
	baseUrl string
}

func WithBaseUrl(baseUrl string) ClientOption {
	return &withBaseUrl{baseUrl: baseUrl}
}

func (o *withBaseUrl) apply(client *Client) {
	client.baseUrl = o.baseUrl
}

func (c *Client) NewRequest(ctx context.Context, method, urlPath string, body io.Reader) (*http.Request, error) {
	reqURL, err := url.Parse(c.baseUrl)
	if err != nil {
		return nil, err
	}
	reqURL, err = reqURL.Parse(urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprint("Bearer ", c.token))
	}
	req.Header.Add("User-Agent", fmt.Sprintf("turso-go/%s (%s/%s)", c.version, runtime.GOOS, runtime.GOARCH))
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Get(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, "GET", path, body)
}

func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, "POST", path, body)
}

func (c *Client) Patch(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, "PATCH", path, body)
}

func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, "PUT", path, body)
}

func (c *Client) Delete(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, "DELETE", path, body)
}

func (c *Client) Upload(ctx context.Context, path string, fileData *os.File) (*http.Response, error) {
	body, bodyWriter := io.Pipe()
	writer := multipart.NewWriter(bodyWriter)
	go func() {
		formFile, err := writer.CreateFormFile("file", fileData.Name())
		if err != nil {
			bodyWriter.CloseWithError(err)
			return
		}
		if _, err := io.Copy(formFile, fileData); err != nil {
			bodyWriter.CloseWithError(err)
			return
		}
		bodyWriter.CloseWithError(writer.Close())
	}()
	req, err := c.NewRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) isNotMemberErr(status int) bool {
	if status == http.StatusForbidden && c.Org != "" {
		return true
	}
	return false
}

func (c *Client) notMemberErr() error {
	return fmt.Errorf("not a member of organization %s", c.Org)
}
