# turso-go

⚠️ This SDK is still in development and is not ready for production use.

## Installation

```sh
go get github.com/alehechka/turso-go
```

## Usage

```go
package main

import (
    "context"
    "github.com/alehechka/turso-go"
)

func main() {
    ctx := context.Background()

    client, err := turso.New("my-token", "my-org")
}
```

The SDK includes a discoverable API for each area available in the Turso Platform API, options include:

```go
client.Instances.List(ctx, "db-name")
client.Databases.List(ctx)
client.Feedback.Submit(ctx, "summary", "feedback")
client.Organizations.List(ctx)
client.ApiTokens.List(ctx)
client.Locations.List(ctx)
client.Tokens.Validate(ctx, "my-token")
client.Users.GetUser(ctx)
client.Plans.List(ctx)
client.Subscriptions.Get(ctx)
client.Billing.Portal(ctx)
client.Groups.List(ctx)
client.Invoices.List(ctx)
```