package factory

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
)

const defaultHost = "github.com"

type Config struct {
	Token               string
	Endpoint            string
	DialTimeout         time.Duration
	TLSHandshakeTimeout time.Duration
	Timeout             time.Duration
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

func NewGitHubClient(opts ...Option) (*github.Client, error) {
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

	token, v3ep := getTokenAndEndpointFromEnv()

	if c.Token == "" {
		c.Token = token
	}

	if c.Token == "" {
		return nil, fmt.Errorf("env %s is not set", "GITHUB_TOKEN")
	}

	if c.Endpoint == "" {
		c.Endpoint = v3ep
	}

	v3c := github.NewClient(httpClient(c))
	if c.Endpoint != "" {
		baseEndpoint, err := url.Parse(c.Endpoint)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(baseEndpoint.Path, "/") {
			baseEndpoint.Path += "/"
		}
		v3c.BaseURL = baseEndpoint
	}

	return v3c, nil
}

func getTokenAndEndpointFromEnv() (string, string) {
	var token, v3ep string
	if os.Getenv("GH_HOST") != "" && os.Getenv("GH_HOST") != defaultHost {
		// GitHub Enterprise Server
		token = os.Getenv("GH_ENTERPRISE_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_ENTERPRISE_TOKEN")
		}
		v3ep = fmt.Sprintf("https://%s/api/v3", os.Getenv("GH_HOST"))
	} else if os.Getenv("GH_TOKEN") != "" {
		// GitHub.com
		token = os.Getenv("GH_TOKEN")
	} else {
		// GitHub Actions
		token = os.Getenv("GITHUB_TOKEN")
		v3ep = os.Getenv("GITHUB_API_URL")
	}
	return token, v3ep
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	return rt.transport.RoundTrip(r)
}

func httpClient(c *Config) *http.Client {
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
