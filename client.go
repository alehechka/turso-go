package turso

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

// Collection of all turso clients
type Client struct {
	baseUrl    *url.URL
	token      string
	Org        string
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

func New(base *url.URL, token string, org string, options ...ClientOption) *Client {
	c := &Client{baseUrl: base, token: token, Org: org, httpClient: http.DefaultClient}

	for _, option := range options {
		option.apply(c)
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
	return c
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

func (c *Client) newRequest(ctx context.Context, method, urlPath string, body io.Reader) (*http.Request, error) {
	url, err := url.Parse(c.baseUrl.String())
	if err != nil {
		return nil, err
	}
	url, err = url.Parse(urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprint("Bearer ", c.token))
	}
	req.Header.Add("User-Agent", fmt.Sprintf("turso-go/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH))
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, path, body)
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
	return c.do(ctx, "GET", path, body)
}

func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, "POST", path, body)
}

func (c *Client) Patch(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, "PATCH", path, body)
}

func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, "PUT", path, body)
}

func (c *Client) Delete(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, "DELETE", path, body)
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
	req, err := c.newRequest(ctx, "POST", path, body)
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
