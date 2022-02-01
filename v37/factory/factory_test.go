package factory

import (
	"context"
	"testing"

	"github.com/google/go-github/v37/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestGetTokenAndEndpoints(t *testing.T) {
	tests := []struct {
		GH_HOST                 string
		GH_ENTERPRISE_TOKEN     string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GITHUB_TOKEN            string
		GITHUB_API_URL          string
		wantToken               string
		wantEndpoint            string
		wantUploadEndpoint      string
		wantV4Endpoint          string
	}{
		{"", "", "", "", "", "", "", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"git.example.com", "", "", "", "", "", "", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GITHUB_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"github.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "GH_TOKEN", "", "", "GH_TOKEN", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "", "GITHUB_TOKEN", "", "GITHUB_TOKEN", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "", "", "GITHUB_API_URL", "", "GITHUB_API_URL", "https://uploads.github.com", "https://api.github.com/graphql"},
	}
	for _, tt := range tests {
		t.Setenv("GH_HOST", tt.GH_HOST)
		t.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
		t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
		t.Setenv("GH_TOKEN", tt.GH_TOKEN)
		t.Setenv("GITHUB_TOKEN", tt.GITHUB_TOKEN)
		t.Setenv("GITHUB_API_URL", tt.GITHUB_API_URL)
		gotToken, gotEndpoint, gotUploadEndpoint, gotV4Endpoint := GetTokenAndEndpoints()

		if gotToken != tt.wantToken {
			t.Errorf("got %v\nwant %v", gotToken, tt.wantToken)
		}

		if gotEndpoint != tt.wantEndpoint {
			t.Errorf("got %v\nwant %v", gotEndpoint, tt.wantEndpoint)
		}

		if gotUploadEndpoint != tt.wantUploadEndpoint {
			t.Errorf("got %v\nwant %v", gotUploadEndpoint, tt.wantUploadEndpoint)
		}

		if gotV4Endpoint != tt.wantV4Endpoint {
			t.Errorf("got %v\nwant %v", gotV4Endpoint, tt.wantV4Endpoint)
		}
	}
}

func TestEndpoint(t *testing.T) {
	tests := []struct {
		GH_HOST                 string
		GH_ENTERPRISE_TOKEN     string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GITHUB_API_URL          string
		wantEndpoint            string
		wantUploadURL           string
	}{
		{"", "", "", "", "", "https://api.github.com/", "https://uploads.github.com/"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "", "", "", "https://git.example.com/api/v3/", "https://git.example.com/api/uploads/"},
		{"", "", "", "", "https://git.example.com/api/v3", "https://git.example.com/api/v3/", "https://git.example.com/api/uploads/"},
		{"", "", "", "", "https://api.github.com", "https://api.github.com/", "https://uploads.github.com/"},
	}
	t.Setenv("GITHUB_TOKEN", "GITHUB_TOKEN")

	for _, tt := range tests {
		t.Setenv("GH_HOST", tt.GH_HOST)
		t.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
		t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
		t.Setenv("GH_TOKEN", tt.GH_TOKEN)
		t.Setenv("GITHUB_API_URL", tt.GITHUB_API_URL)

		client, err := NewGithubClient()
		if err != nil {
			t.Fatal(err)
			continue
		}

		gotEndpoint := client.BaseURL.String()
		gotUploadURL := client.UploadURL.String()

		if gotEndpoint != tt.wantEndpoint {
			t.Errorf("got %v\nwant %v", gotEndpoint, tt.wantEndpoint)
		}

		if gotUploadURL != tt.wantUploadURL {
			t.Errorf("got %v\nwant %v", gotUploadURL, tt.wantUploadURL)
		}
	}
}

func TestNewGithubClient(t *testing.T) {
	_, err := NewGithubClient()
	if err != nil {
		t.Error(err)
	}
}

func TestNewGithubClientUsingMock(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUsersByUsername,
			github.User{
				Name: github.String("foobar"),
			},
		),
	)
	c, err := NewGithubClient(HTTPClient(mockedHTTPClient))
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
