package factory

import (
	"context"
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
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/google/go-github/v67/github"
)

const defaultHost = "github.com"
const defaultV3Endpoint = "https://api.github.com"
const defaultUploadEndpoint = "https://uploads.github.com"
const defaultV4Endpoint = "https://api.github.com/graphql"

type Config struct {
	Token               string
	Endpoint            string
	Owner               string
	Repo                string
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

func Owner(owner string) Option {
	return func(c *Config) error {
		c.Owner = owner
		return nil
	}
}

func OwnerRepo(ownerrepo string) Option {
	return func(c *Config) error {
		splitted := strings.Split(ownerrepo, "/")
		if len(splitted) != 2 {
			return errors.New("invalid owner/repo format")
		}
		c.Owner = splitted[0]
		c.Repo = splitted[1]
		return nil
	}
}

// NewGithubClient returns github.com/google/go-github/v67/github.Client with environment variable resolution.
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
func GetTokenAndEndpoints() (token string, v3ep string, v3upload string, v4ep string) {
	token, v3ep, v3upload, v4ep, _, _, _ = GetAllDetected()
	return token, v3ep, v3upload, v4ep
}

// GetAllDetected returns token, endpoints, host and sources. The endpoints to be generated are URLs without a trailing slash.
func GetAllDetected() (token, v3ep, v3upload, v4ep, host, hostSource, tokenSource string) {
	host, hostSource = auth.DefaultHost()
	token, tokenSource = auth.TokenForHost(host)
	v3ep = defaultV3Endpoint
	v3upload = defaultUploadEndpoint
	v4ep = defaultV4Endpoint
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

	return token, v3ep, v3upload, v4ep, host, hostSource, tokenSource
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
	if envAppID == "" || envPrivateKey == "" {
		return nil, errors.New("not enough credentials to authenticate using GitHub app")
	}
	appID, err := strconv.ParseInt(envAppID, 10, 64)
	if err != nil {
		return nil, err
	}
	privateKey := []byte(repairKey(envPrivateKey))
	var installationID int64
	if envInstallaitonID == "" {
		installationID, err = detectInstallationID(c, appID, privateKey, ep)
		if err != nil {
			return nil, err
		}
	} else {
		installationID, err = strconv.ParseInt(envInstallaitonID, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	hc := httpClient(c)
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
	if err != nil {
		return nil, err
	}
	itr.BaseURL = ep
	hc.Transport = itr
	return hc, nil
}

func detectInstallationID(c *Config, appID int64, privateKey []byte, ep string) (int64, error) {
	owner, repo, err := detectOwnerRepo(c)
	if err != nil {
		return 0, err
	}
	tr := http.DefaultTransport
	atr, err := ghinstallation.NewAppsTransport(tr, appID, privateKey)
	if err != nil {
		return 0, err
	}
	atr.BaseURL = ep
	hc := &http.Client{Transport: atr}
	gc := github.NewClient(hc)
	baseEndpoint, err := url.Parse(ep)
	if err != nil {
		return 0, err
	}
	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}
	gc.BaseURL = baseEndpoint
	ctx := context.Background()
	if repo != "" {
		i, _, err := gc.Apps.FindRepositoryInstallation(ctx, owner, repo)
		if err != nil {
			return 0, err
		}
		return i.GetID(), nil
	}
	page := 0
	for {
		is, res, err := gc.Apps.ListInstallations(context.Background(), &github.ListOptions{Page: page, PerPage: 1000})
		if err != nil {
			return 0, err
		}
		for _, i := range is {
			if owner == i.GetAccount().GetLogin() {
				return i.GetID(), nil
			}
		}
		if res.NextPage >= res.LastPage {
			break
		}
		page = res.NextPage
	}
	return 0, fmt.Errorf("could not installation for %s", owner)
}

func detectOwnerRepo(c *Config) (string, string, error) {
	if c.Owner != "" {
		return c.Owner, c.Repo, nil
	}
	if hostownerrepo := os.Getenv("GH_REPO"); hostownerrepo != "" {
		splitted := strings.Split(hostownerrepo, "/")
		switch {
		case len(splitted) < 2:
			return "", "", fmt.Errorf("invalid env GH_REPO: %s", hostownerrepo)
		case len(splitted) == 3:
			return splitted[1], splitted[2], nil
		default:
			return splitted[0], splitted[1], nil
		}
	}
	if ownerrepo := os.Getenv("GITHUB_REPOSITORY"); ownerrepo != "" {
		splitted := strings.Split(ownerrepo, "/")
		if len(splitted) < 2 {
			return "", "", fmt.Errorf("invalid env GITHUB_REPOSITORY: %s", ownerrepo)
		}
		return splitted[0], splitted[1], nil
	}
	if owner := os.Getenv("GITHUB_REPOSITORY_OWNER"); owner != "" {
		return owner, "", nil
	}
	return "", "", errors.New("could not detect repository")
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
