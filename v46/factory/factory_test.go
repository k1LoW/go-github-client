package factory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/google/go-github/v46/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestGetTokenAndEndpoints(t *testing.T) {
	t.Setenv("GH_CONFIG_DIR", filepath.Join(testdataDir(t), "config"))
	// set config
	// ref: https://github.com/cli/go-gh/blob/98bbeb261673e1c506e965ab1553bfbaf5318250/pkg/config/config.go#L124-L128
	_, _ = config.Read(&config.Config{})

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
		{"", "", "", "", "", "", "gho_XXXXXxxxxXXXXxxxXXXXXX", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"git.example.com", "", "", "", "", "", "", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"git.example.com", "", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GITHUB_ENTERPRISE_TOKEN", "https://git.example.com/api/v3", "https://git.example.com/api/uploads", "https://git.example.com/api/graphql"},
		{"github.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "gho_XXXXXxxxxXXXXxxxXXXXXX", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "GH_TOKEN", "", "", "GH_TOKEN", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "", "GITHUB_TOKEN", "", "GITHUB_TOKEN", "https://api.github.com", "https://uploads.github.com", "https://api.github.com/graphql"},
		{"", "", "", "", "", "GITHUB_API_URL", "gho_XXXXXxxxxXXXXxxxXXXXXX", "GITHUB_API_URL", "https://uploads.github.com", "https://api.github.com/graphql"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
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
		})
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
	user, _, err := c.Users.Get(ctx, "foobar")
	if err != nil {
		t.Fatal(err)
	}
	got := user.GetName()
	if want := "foobar"; got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestDetectOwnerRepo(t *testing.T) {
	tests := []struct {
		GH_REPO                 string
		GITHUB_REPOSITORY       string
		GITHUB_REPOSITORY_OWNER string
		wantOwner               string
		wantRepo                string
		wantErr                 bool
	}{
		{"", "", "", "", "", true},
		{"example/myrepo", "", "", "example", "myrepo", false},
		{"git.example.com/example/myrepo", "", "", "example", "myrepo", false},
		{"", "example/myrepo", "", "example", "myrepo", false},
		{"example/ourrepo", "example/myrepo", "", "example", "ourrepo", false},
		{"", "", "example", "example", "", false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Setenv("GH_REPO", tt.GH_REPO)
			t.Setenv("GITHUB_REPOSITORY", tt.GITHUB_REPOSITORY)
			t.Setenv("GITHUB_REPOSITORY_OWNER", tt.GITHUB_REPOSITORY_OWNER)
			gotOwner, gotRepo, err := detectOwnerRepo()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if gotOwner != tt.wantOwner {
				t.Errorf("got %v\nwant %v", gotOwner, tt.wantOwner)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("got %v\nwant %v", gotRepo, tt.wantRepo)
			}
		})
	}
}

func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
