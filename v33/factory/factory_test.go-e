package factory

import (
	"testing"
)

func TestGetTokenAndEndpointFromEnv(t *testing.T) {
	tests := []struct {
		GH_HOST                 string
		GH_ENTERPRISE_TOKEN     string
		GITHUB_ENTERPRISE_TOKEN string
		GH_TOKEN                string
		GITHUB_TOKEN            string
		GITHUB_API_URL          string
		wantToken               string
		wantEndpoint            string
	}{
		{"", "", "", "", "", "", "", ""},
		{"git.example.com", "", "", "", "", "", "", "https://git.example.com/api/v3"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3"},
		{"git.example.com", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GH_ENTERPRISE_TOKEN", "https://git.example.com/api/v3"},
		{"git.example.com", "", "GITHUB_ENTERPRISE_TOKEN", "", "", "", "GITHUB_ENTERPRISE_TOKEN", "https://git.example.com/api/v3"},
		{"github.com", "GH_ENTERPRISE_TOKEN", "", "", "", "", "", ""},
		{"", "", "", "GH_TOKEN", "", "", "GH_TOKEN", ""},
		{"", "", "", "", "GITHUB_TOKEN", "", "GITHUB_TOKEN", ""},
		{"", "", "", "", "", "GITHUB_API_URL", "", "GITHUB_API_URL"},
	}
	for _, tt := range tests {
		t.Setenv("GH_HOST", tt.GH_HOST)
		t.Setenv("GH_ENTERPRISE_TOKEN", tt.GH_ENTERPRISE_TOKEN)
		t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.GITHUB_ENTERPRISE_TOKEN)
		t.Setenv("GH_TOKEN", tt.GH_TOKEN)
		t.Setenv("GITHUB_TOKEN", tt.GITHUB_TOKEN)
		t.Setenv("GITHUB_API_URL", tt.GITHUB_API_URL)
		gotToken, gotEndpoint := getTokenAndEndpointFromEnv()

		if gotToken != tt.wantToken {
			t.Errorf("got %v\nwant %v", gotToken, tt.wantToken)
		}

		if gotEndpoint != tt.wantEndpoint {
			t.Errorf("got %v\nwant %v", gotEndpoint, tt.wantEndpoint)
		}
	}
}

func TestNewGitHubClient(t *testing.T) {
	_, err := NewGitHubClient()
	if err != nil {
		t.Error(err)
	}
}
