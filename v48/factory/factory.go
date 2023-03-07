package factory

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/cli/go-gh/pkg/auth"
	"github.com/google/go-github/v48/github"
)

const defaultHost = "github.com"
const defaultV3Endpoint = "https://api.github.com"
const defaultUploadEndpoint = "https://uploads.github.com"
const defaultV4Endpoint = "https://api.github.com/graphql"

type Config struct {
	Token               string
	Endpoint            string
	DialTimeout         time.Duration
	TLSHandshakeTimeout time.Duration
	Timeout             time.Duration
	HTTPClient          *http.Client
	SkipAuth            bool
}

type Option func(*Config) error

func Token(t string) Option {
	return func(c *Config) error {
		if t != "" {
			c.Token = t
		}
		return nil
	}
}

func Endpoint(t string) Option {
	return func(c *Config) error {
		if t != "" {
			c.Endpoint = t
		}
		return nil
	}
}

func DialTimeout(to time.Duration) Option {
	return func(c *Config) error {
		if to > 0 {
			c.DialTimeout = to
		}
		return nil
	}
}

func TLSHandshakeTimeout(to time.Duration) Option {
	return func(c *Config) error {
		if to > 0 {
			c.TLSHandshakeTimeout = to
		}
		return nil
	}
}

func Timeout(to time.Duration) Option {
	return func(c *Config) error {
		if to > 0 {
			c.Timeout = to
		}
		return nil
	}
}

func HTTPClient(httpClient *http.Client) Option {
	return func(c *Config) error {
		if httpClient != nil {
			c.HTTPClient = httpClient
		}
		return nil
	}
}

func SkipAuth(enable bool) Option {
	return func(c *Config) error {
		c.SkipAuth = enable
		return nil
	}
}

// NewGithubClient returns github.com/google/go-github/v48/github.Client with environment variable resolution
func NewGithubClient(opts ...Option) (*github.Client, error) {
	c := &Config{
		Token:               "",
		DialTimeout:         5 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		Timeout:             30 * time.Second,
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	token, v3ep, v3upload, _ := GetTokenAndEndpoints()

	if c.Token == "" {
		c.Token = token
	}

	if c.SkipAuth {
		c.Token = ""
	}

	ep := c.Endpoint
	if ep == "" {
		ep = v3ep
	}

	if !c.SkipAuth && c.Token == "" && c.HTTPClient == nil {
		hc, err := newHTTPClientUsingGitHubApp(c, ep)
		if err != nil {
			return nil, errors.New("no credentials found")
		}
		c.HTTPClient = hc
	}

	v3c := github.NewClient(httpClient(c))
	baseEndpoint, err := url.Parse(ep)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}
	v3c.BaseURL = baseEndpoint

	if c.Endpoint != "" {
		if !strings.Contains(baseEndpoint.Host, defaultHost) {
			v3c.UploadURL, err = url.Parse(fmt.Sprintf("https://%s/api/uploads/", baseEndpoint.Host))
			if err != nil {
				return nil, err
			}
		}
	} else {
		uploadEndpoint, err := url.Parse(v3upload)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(uploadEndpoint.Path, "/") {
			uploadEndpoint.Path += "/"
		}
		v3c.UploadURL = uploadEndpoint
	}

	return v3c, nil
}

// GetTokenAndEndpoints returns token and endpoints. The endpoints to be generated are URLs without a trailing slash.
func GetTokenAndEndpoints() (string, string, string, string) {
	host, _ := auth.DefaultHost()
	token, _ := auth.TokenForHost(host)
	v3ep := defaultV3Endpoint
	v3upload := defaultUploadEndpoint
	v4ep := defaultV4Endpoint
	if host != defaultHost {
		// GitHub Enterprise Server
		v3ep = fmt.Sprintf("https://%s/api/v3", host)
		v3upload = fmt.Sprintf("https://%s/api/uploads", host)
		v4ep = fmt.Sprintf("https://%s/api/graphql", host)
	} else {
		// GitHub Actions or GitHub.com
		if os.Getenv("GITHUB_API_URL") != "" {
			v3ep = os.Getenv("GITHUB_API_URL")
			ep, err := url.Parse(v3ep)
			if err == nil && ep.Host != "" {
				if !strings.Contains(ep.Host, defaultHost) {
					v3upload = fmt.Sprintf("https://%s/api/uploads", ep.Host)
				}
			}
		}
		if os.Getenv("GITHUB_GRAPHQL_URL") != "" {
			v4ep = os.Getenv("GITHUB_GRAPHQL_URL")
		}
	}

	return token, v3ep, v3upload, v4ep
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if rt.accessToken != "" {
		r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	}
	return rt.transport.RoundTrip(r)
}

func newHTTPClientUsingGitHubApp(c *Config, ep string) (*http.Client, error) {
	envAppID := os.Getenv("GITHUB_APP_ID")
	envInstallaitonID := os.Getenv("GITHUB_APP_INSTALLATION_ID")
	envPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if envAppID == "" || envInstallaitonID == "" || envPrivateKey == "" {
		// TODO: detect installationID
		return nil, errors.New("not enough credentials to authenticate using GitHub app")
	}
	appID, err := strconv.ParseInt(envAppID, 10, 64)
	if err != nil {
		return nil, err
	}
	installationID, err := strconv.ParseInt(envInstallaitonID, 10, 64)
	if err != nil {
		return nil, err
	}
	privateKey := []byte(repairKey(envPrivateKey))
	hc := httpClient(c)
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
	if err != nil {
		return nil, err
	}
	itr.BaseURL = ep
	hc.Transport = itr
	return hc, nil
}

func httpClient(c *Config) *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: c.DialTimeout,
		}).Dial,
		TLSHandshakeTimeout: c.TLSHandshakeTimeout,
	}
	rt := roundTripper{
		transport:   t,
		accessToken: c.Token,
	}
	return &http.Client{
		Timeout:   c.Timeout,
		Transport: rt,
	}
}

func repairKey(in string) string {
	repairRep := strings.NewReplacer("-----BEGIN OPENSSH PRIVATE KEY-----", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", "-----BEGIN RSA PRIVATE KEY-----", "-----BEGIN_RSA_PRIVATE_KEY-----", "-----END RSA PRIVATE KEY-----", "-----END_RSA_PRIVATE_KEY-----", " ", "\n", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----BEGIN OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----", "-----BEGIN_RSA_PRIVATE_KEY-----", "-----BEGIN RSA PRIVATE KEY-----", "-----END_RSA_PRIVATE_KEY-----", "-----END RSA PRIVATE KEY-----")
	return repairRep.Replace(repairRep.Replace(in))
}
