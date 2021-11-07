package main

import (
	"context"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/k1LoW/go-github-client/v39/factory"
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
