package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

// TODO: Remove "user"
func (i integration) NewClient(ctx context.Context) (*github.Client, error) {
	// Extract the connection token from the given context.
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := i.vars.Reveal(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}
	if pat := data.Get(vars.PAT); pat.IsValid() {
		return github.NewTokenClient(ctx, pat.Value()), nil
	} else {
		return i.newClientWithInstallJWT(data)
	}
}

// NewClientWithInstallJWT returns a GitHub client that
// uses a newly-generated GitHub app installation JWT.
func (i integration) newClientWithInstallJWT(data sdktypes.Vars) (*github.Client, error) {
	// Initialize and return a GitHub client with a JWT.
	s := data.GetValue(vars.AppID)
	if s == "" {
		return nil, fmt.Errorf("app ID not found")
	}
	aid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID %q", s)
	}

	s = data.GetValue(vars.InstallID)
	if s == "" {
		return nil, fmt.Errorf("install ID not found")
	}
	iid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid install ID %q", s)
	}

	return i.newClientWithInstallJWTFromGitHubIDs(aid, iid)
}

// newClientWithInstallJWTFromGitHubIDs generates a GitHub app
// installation JWT based on the given GitHub app ID and installation ID. See:
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
func (i integration) newClientWithInstallJWTFromGitHubIDs(appID, installID int64) (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport
	enterpriseURL, err := enterpriseURL()
	if err != nil {
		return nil, err
	}

	// Wrap the shared transport.
	itr, err := ghinstallation.New(tr, appID, installID, getPrivateKey())
	if err != nil {
		return nil, err
	}
	if enterpriseURL != "" {
		itr.BaseURL = enterpriseURL + "/api/v3"
	}

	// Initialize a client with the generated JWT injected into outbound requests.
	client := github.NewClient(&http.Client{Transport: itr})
	if enterpriseURL != "" {
		client, err = client.WithEnterpriseURLs(enterpriseURL, enterpriseURL)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (i integration) newAnonymousClient() (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	enterpriseURL, err := enterpriseURL()
	if err != nil {
		return nil, err
	}

	// Initialize a client with the generated JWT injected into outbound requests.
	client := github.NewClient(nil)
	if enterpriseURL != "" {
		client, err = client.WithEnterpriseURLs(enterpriseURL, enterpriseURL)
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
