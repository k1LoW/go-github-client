# go-github-client

## Usage

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
