# go-github-client

:octocat: [go-github](https://github.com/google/go-github) client factory.

## Usage

`go-github-client/[VERSION]/factory.NewGithubClient()` returns `github.com/google/go-github/[VERSION]/github.Client` with environment variable resolution

``` go
package main

import (
	"context"
	"fmt"

	"github.com/k1LoW/go-github-client/v39/factory"
)

func main() {
	ctx := context.Background()
	c, _ := factory.NewGithubClient()
	u, _, _ := c.Users.Get(ctx, "k1LoW")
	fmt.Printf("%s\n", u.GetLocation())
}
```
