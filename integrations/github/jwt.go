package github

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

// newClientWithInstallJWT returns a GitHub client that
// uses a newly-generated GitHub app installation JWT.
func newClientWithInstallJWT(data sdktypes.Vars) (*github.Client, error) {
	// Initialize and return a GitHub client with a JWT.
	s := data.GetValue(vars.AppID)
	if s == "" {
		return nil, errors.New("app ID not found")
	}
	aid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID %q", s)
	}

	s = data.GetValue(vars.InstallID)
	if s == "" {
		return nil, errors.New("install ID not found")
	}
	iid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid install ID %q", s)
	}

	return newClientWithInstallJWTFromGitHubIDs(aid, iid, data.GetValue(vars.PrivateKey))
}

// newClientWithInstallJWTFromGitHubIDs generates a GitHub app
// installation JWT based on the given GitHub app ID and installation ID. See:
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
// The private key is used to sign the JWT and determine whether this is a
// user-defined GitHub App or a GitHub App.
func newClientWithInstallJWTFromGitHubIDs(appID, installID int64, privateKey string) (*github.Client, error) {
	client, err := NewClientFromGitHubAppID(appID, privateKey)
	if err != nil {
		return nil, err
	}

	atr := client.Client().Transport.(*ghinstallation.AppsTransport)
	itr := ghinstallation.NewFromAppsTransport(atr, installID)
	client = github.NewClient(&http.Client{Transport: itr})

	enterpriseURL, err := enterpriseURL()
	if err != nil {
		return nil, err
	}
	if enterpriseURL != "" {
		client, err = client.WithEnterpriseURLs(enterpriseURL, enterpriseURL)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// NewClientFromGitHubAppID generates a GitHub app JWT based on its ID. The private key
// determines whether this is a user-defined GitHub App and is used to sign the JWT.
// If the private key is not provided, the environment variable is used.
func NewClientFromGitHubAppID(appID int64, privateKey string) (*github.Client, error) {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport.
	atr, err := ghinstallation.NewAppsTransport(tr, appID, getPrivateKey(privateKey))
	if err != nil {
		return nil, err
	}

	// Initialize a client with the generated JWT injected into outbound requests.
	client := github.NewClient(&http.Client{Transport: atr})

	enterpriseURL, err := enterpriseURL()
	if err != nil {
		return nil, err
	}
	if enterpriseURL != "" {
		client, err = client.WithEnterpriseURLs(enterpriseURL, enterpriseURL)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func getPrivateKey(privateKey string) []byte {
	// Check if this is a custom OAuth connection
	if privateKey != "" {
		return []byte(strings.ReplaceAll(privateKey, "\\n", "\n"))
	}
	s, ok := os.LookupEnv(privateKeyEnvVar)
	if ok {
		return []byte(strings.ReplaceAll(s, "\\n", "\n"))
	}
	// This is solely for unit tests. It's safe dead code in production because
	// in production we check that the environment variable exists.
	k, err := rsa.GenerateKey(rand.Reader, 2048)
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
