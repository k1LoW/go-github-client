# go-github-client

:octocat: [go-github](https://github.com/google/go-github) client factory.

## Usage

`go-github-client/[VERSION]/factory.NewGithubClient()` returns `github.com/google/go-github/[VERSION]/github.Client` with environment variable resolution

``` go
package main

import (
	"context"
	"fmt"

	"github.com/k1LoW/go-github-client/v49/factory"
)

func main() {
	ctx := context.Background()
	c, _ := factory.NewGithubClient()
	u, _, _ := c.Users.Get(ctx, "k1LoW")
	fmt.Printf("%s\n", u.GetLocation())
}
```

### Mocking

``` go
package main

import (
	"context"
	"testing"

	"github.com/google/go-github/v49/github"
	"github.com/k1LoW/go-github-client/v49/factory"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestUsingMock(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{
				Name: github.String("foobar"),
			},
		),
	)
	c, err := factory.NewGithubClient(factory.HTTPClient(mockedHTTPClient))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	user, _, err := c.Users.Get(ctx, "myuser")
	if err != nil {
		t.Fatal(err)
	}
	got := user.GetName()
	if want := "foobar"; got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}
```

## Environment variables that affect client initialization

- `GH_TOKEN`, `GITHUB_TOKEN`
- `GH_ENTERPRISE_TOKEN`, `GITHUB_ENTERPRISE_TOKEN`
- `GH_HOST`, `GITHUB_API_URL`, `GITHUB_GRAPHQL_URL`
- `GH_CONFIG_DIR`
- `GITHUB_APP_ID`, `GITHUB_APP_INSTALLATION_ID`, `GITHUB_APP_PRIVATE_KEY`, `GH_REPO`, `GITHUB_REPOSITORY`, `GITHUB_REPOSITORY_OWNER` for authentication with a GitHub App

## Versioning

| Version | Description |
| --- | --- |
| Major | google/go-github major version |
| Minor | google/go-github minor version |
| Patch | google/go-github patch version + k1LoW/go-github-client update |
