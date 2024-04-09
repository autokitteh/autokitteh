package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"strconv"
	"strings"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
)

const (
	// privateKeyEnvVar is the name of an environment variable that contains a
	// SECRET PEM-encoded GitHub private key which is required to sign JWTs.
	privateKeyEnvVar = "GITHUB_PRIVATE_KEY"

	// enterpriseURLEnvVar is the name of an environment variable that contains
	// the (cloud or on-prem) base URL of a GitHub Enterprise Server instance.
	// This URL should not have a path suffix like "/api/v3" or "/api/uploads",
	// autokitteh will append those as needed.
	enterpriseURLEnvVar = "GITHUB_ENTERPRISE_URL"
)

func (i integration) NewClient(ctx context.Context) (*github.Client, error) {
	data, err := i.getConnection(ctx)
	if err != nil {
		return nil, err
	}
	if pat, ok := data["PAT"]; ok {
		return github.NewTokenClient(ctx, pat), nil
	} else {
		return i.NewClientWithInstallJWT(ctx)
	}
}

// NewClientWithAppJWT returns a GitHub client
// that uses a newly-generated GitHub app JWT.
func (i integration) NewClientWithAppJWT(ctx context.Context) (*github.Client, error) {
	data, err := i.getConnection(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize and return a GitHub client with a JWT.
	aid, err := strconv.ParseInt(data["appID"], 10, 64)
	if err != nil {
		return nil, err
	}
	return i.NewClientWithAppJWTFromGitHubID(aid)
}

// NewClientWithInstallJWT returns a GitHub client that
// uses a newly-generated GitHub app installation JWT.
func (i integration) NewClientWithInstallJWT(ctx context.Context) (*github.Client, error) {
	data, err := i.getConnection(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize and return a GitHub client with a JWT.
	aid, err := strconv.ParseInt(data["appID"], 10, 64)
	if err != nil {
		return nil, err
	}
	iid, err := strconv.ParseInt(data["installID"], 10, 64)
	if err != nil {
		return nil, err
	}
	return i.NewClientWithInstallJWTFromGitHubIDs(aid, iid)
}

// getConnection calls the Get method in SecretsService.
func (i integration) getConnection(ctx context.Context) (map[string]string, error) {
	// Extract the connection token from the given context.
	cfg := sdkmodule.FunctionDataFromContext(ctx)
	if cfg == nil {
		cfg = []byte{}
	}

	c, err := i.secrets.Get(context.Background(), i.scope, string(cfg))
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewClientWithAppJWTFromGitHubID generates a GitHub app JWT based on the
// given GitHub app ID, and returns a GitHub client that uses it. Based on:
//   - https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app
//   - https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app
func (i integration) NewClientWithAppJWTFromGitHubID(appID int64) (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport
	u, err := enterpriseURL()
	if err != nil {
		return nil, err
	}

	// Wrap the shared transport.
	atr, err := ghinstallation.NewAppsTransport(tr, appID, getPrivateKey())
	if err != nil {
		return nil, err
	}
	if u != "" {
		atr.BaseURL = u + "/api/v3"
	}

	// Initialize a client with the generated JWT injected into outbound requests.
	client := github.NewClient(&http.Client{Transport: atr})
	if u != "" {
		client, err = client.WithEnterpriseURLs(u, u)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// NewClientWithInstallJWTFromGitHubIDs generates a GitHub app
// installation JWT based on the given GitHub app ID and installation ID. See:
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
func (i integration) NewClientWithInstallJWTFromGitHubIDs(appID, installID int64) (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport
	u, err := enterpriseURL()
	if err != nil {
		return nil, err
	}

	// Wrap the shared transport.
	itr, err := ghinstallation.New(tr, appID, installID, getPrivateKey())
	if err != nil {
		return nil, err
	}
	if u != "" {
		itr.BaseURL = u + "/api/v3"
	}

	// Initialize a client with the generated JWT injected into outbound requests.
	client := github.NewClient(&http.Client{Transport: itr})
	if u != "" {
		client, err = client.WithEnterpriseURLs(u, u)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func getPrivateKey() []byte {
	s, ok := os.LookupEnv(privateKeyEnvVar)
	if ok {
		return []byte(strings.ReplaceAll(s, "\\n", "\n"))
	}
	// This is solely for unit tests. It's safe dead code in production because
	// in production we check that the environment variable exists.
	k, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil
	}
	b := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k),
	}
	return pem.EncodeToMemory(b)
}

func enterpriseURL() (string, error) {
	u := os.Getenv(enterpriseURLEnvVar)
	if u == "" {
		return u, nil
	}

	return kittehs.NormalizeURL(u, true)
}
