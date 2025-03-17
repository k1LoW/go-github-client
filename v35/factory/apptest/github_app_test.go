package apptest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/k1LoW/go-github-client/v35/factory"
	"github.com/k1LoW/httpstub"
)

const (
	testAppID          = 1
	testInstallationID = 2
	//nolint:gosec
	testPrivateKey     = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAxN+b3zon7HKP3vd7wxMpXfugseMzk7NAdYzq23Ipv9MC/CPD
gaPFjH0vFVoBdn5JL7XqUjOM2qbYdkjHWK4+PACntKaHhOkb22HBl9N5evpibLMM
hIW+dCAYv9+B+/fjNx+2gCTv5WpUDrVw9PNDKdmhqCGsfD3uWR8/iRogwr0nlB5Q
lYuEY6yK14hhBicPhRRW1Guko5Sq6Mxd6m+ua22xqjHNwoKj2rApZhUXcShYjNIK
Enm/2mIDLSzTUtu67cAHMnXEQjtEfVL8Ks1mrz7zkL62qA3jisRcGp5T7DfUkC+O
WiesJ9CQEwE+OxlSgz0wGbIJn/+Pnyc6gLLjfQIDAQABAoIBAHHeq/dXWdQnBxP9
rPXN1XVony+ErEZXvYbANO8sfv1WfTl9Lg2DvjVeCqec4Y+5x3bzD07wRh4JttXj
jnm6foCSGG4ii+vSMKyZRDIevPrma5tXjHvyJ5BfKDGCg1pLrH4rt5EyzBazg17m
jyj+svA30oq+v1c1MvEVY9hW5m/7mHQ9vkhEypZrGgXAXj0H395TaxQHFXLP6nOL
qMCofm+vRYKe3glVP2EaeNraSiTuDSqBJNA5B5w78+F94O7oUfh169YLhuaMa+47
wEnhAnze6DvNVRXAse06EwRAAzL/vLdnjZoCS3TpTqi2XAy4x3ZCalhVgfykrk6b
GShGAvkCgYEA5LM/Vu6e7ggH3u7vPX7kiu4DcvEp+vWpzjvYyGnOHfSHR71T/zce
4tniSKNjOlPvGLzlKByl9ZAVA5y6sMeEQM9DX3UGRRZ57H4yJsuiCxDt4uOC14VM
70H0AgoNYE8KnimcGar9pIVVNYrWp32yu1/mWhCKGO0V2UMp0Rp0qGcCgYEA3F/I
R51r4szLVuAqyS72kzCelWpMolULhvMewCtTS/t17p6k8AAOJ5ZQtT7G2NZx84+W
xBmwkJfHymCGN5VCziQ29OSK+jFrVz/Q3jHDZCtSQoNZ7rO9igLY3BF9wwI0EAF8
sIYMnW8KqrW7gFZ4mBcMb1YUdfjt4S6Ag2C19nsCgYBpG5h4s6KHc1lqtBVwBemz
kEA1i3DnzhAEoKy5LydzzPZ/mhwIp6SiTdEZ4T2xiPHSRL5s+P2tJlMCHf4PUSMP
RjKIpJgFGJdggX87JUuMGnO6WyW/N5xsObuTVFthb/JJToZXpaZ8/mpy+SQ+Rh7m
zuRncEKHwi7Qc3W8jJQg8QKBgQCZe2w11IHrN8729rFV5Qt+gAIy9hHhjXG1z2W/
WW1uIfiE9KDTNnalQ596W/qJ0vESPRM4CNxcGBnh7VANLjuU7swHy5Svo/OqlJuX
5Pi8rx9fi7P699wuXsVCoDwCsWopK5/4IaRvkYLQWjn4rEDZTFQwxrcBYxnqF0US
Oy0AOQKBgQCxnA1jlVNvrIuHejhgUVNB+0iihgtIMM1j/UgMIk4zDaaen6K9USP5
c4pNuVzgPw4B5cRgOviicAgcn02c8JM+gSZVtKf2+Y2EqHJ5Uzl5fOyxQZkyUH9+
S/v7bzqbRiyelnXleEA/SwS6VVjm3PNad6t/iLtENBoPMrBxq3UjLg==
-----END RSA PRIVATE KEY-----
`
	testOwner = "example"
	testRepo  = "myapp"
)

func TestAuthUsingGitHubApp(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GH_ENTERPRISE_TOKEN", "")
	t.Setenv("GITHUB_APP_ID", strconv.Itoa(testAppID))
	t.Setenv("GITHUB_APP_INSTALLATION_ID", strconv.Itoa(testInstallationID))
	t.Setenv("GITHUB_APP_PRIVATE_KEY", testPrivateKey)
	t.Setenv("GH_CONFIG_DIR", "/tmp")
	r := httpstub.NewRouter(t)
	r.Method(http.MethodPost).Path(fmt.Sprintf("/app/installations/%d/access_tokens", testInstallationID)).ResponseString(http.StatusOK, `{}`)
	r.Method(http.MethodGet).Path(fmt.Sprintf("/users/%s/repos", testOwner)).ResponseString(http.StatusOK, `[]`)
	ts := r.Server()
	t.Cleanup(func() {
		ts.Close()
	})
	t.Setenv("GITHUB_API_URL", ts.URL)
	t.Run("t", func(t *testing.T) {
		t.Parallel() // to set GH_CONFIG_DIR and create new config
		c, err := factory.NewGithubClient()
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := c.Repositories.List(context.Background(), testOwner, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestAuthUsingGitHubAppNoInstallationID(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GH_ENTERPRISE_TOKEN", "")
	t.Setenv("GITHUB_APP_ID", strconv.Itoa(testAppID))
	t.Setenv("GITHUB_APP_PRIVATE_KEY", testPrivateKey)
	t.Setenv("GITHUB_REPOSITORY", fmt.Sprintf("%s/%s", testOwner, testRepo))
	t.Setenv("GH_CONFIG_DIR", "/tmp")
	r := httpstub.NewRouter(t)
	r.Method(http.MethodGet).Path(fmt.Sprintf("/repos/%s/%s/installation", testOwner, testRepo)).ResponseString(http.StatusOK, fmt.Sprintf(`{"id": %d}`, testInstallationID))
	r.Method(http.MethodPost).Path(fmt.Sprintf("/app/installations/%d/access_tokens", testInstallationID)).ResponseString(http.StatusOK, `{}`)
	r.Method(http.MethodGet).Path(fmt.Sprintf("/users/%s/repos", testOwner)).ResponseString(http.StatusOK, `[]`)
	ts := r.Server()
	t.Cleanup(func() {
		ts.Close()
	})
	t.Setenv("GITHUB_API_URL", ts.URL)
	t.Run("t", func(t *testing.T) {
		t.Parallel() // to set GH_CONFIG_DIR and create new config
		c, err := factory.NewGithubClient()
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := c.Repositories.List(context.Background(), testOwner, nil); err != nil {
			t.Error(err)
		}
	})
}
