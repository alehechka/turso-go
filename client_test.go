package turso_test

import (
	"context"
	"testing"

	"github.com/alehechka/turso-go"
)

func Test_NewClient_ReturnsErrorOnMissingToken(t *testing.T) {
	_, err := turso.New("", "")
	if err == nil {
		t.Fatal("expected client initialization to return error")
	}
	if err != turso.ErrMissingAPIToken {
		t.Fatalf("expected client initialization to return: %s", turso.ErrMissingAPIToken)
	}
}

func Test_NewClient_ReturnsErrorOnMissingBaseURL(t *testing.T) {
	_, err := turso.New("my-token", "", turso.WithBaseUrl(""))
	if err == nil {
		t.Fatal("expected client initialization to return error")
	}
	if err != turso.ErrMissingBaseURL {
		t.Fatalf("expected client initialization to return: %s", turso.ErrMissingBaseURL)
	}
}

func Test_NewClient_ReturnsErrorOnMissingHTTPClient(t *testing.T) {
	_, err := turso.New("my-token", "", turso.WithHTTPClient(nil))
	if err == nil {
		t.Fatal("expected client initialization to return error")
	}
	if err != turso.ErrMissingHTTPClient {
		t.Fatalf("expected client initialization to return: %s", turso.ErrMissingHTTPClient)
	}
}

func Test_Request_T(t *testing.T) {
	client, err := turso.New("my-token", "my-org")
	if err != nil {
		t.Fatalf(err.Error())
	}

	res, err := client.Databases.List(context.TODO())
	if err == nil {
		t.Fatal("expected request to return error")
	}
	if err.Error() != "failed to get database listing: token contains an invalid number of segments" {
		t.Fatalf("expected certain response, got: %s", err)
	}
	if res != nil {
		t.Fatal("expected response to be nil")
	}
}
