package common

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/internal/xdg"
)

var credsPath = filepath.Join(xdg.ConfigHomeDir(), "credentials")

func readCreds() (creds map[string]string, err error) {
	creds = map[string]string{}

	var bs []byte
	if bs, err = os.ReadFile(credsPath); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}

		return
	}

	if len(bs) != 0 {
		err = yaml.Unmarshal(bs, &creds)
	}

	return
}

func credsHost() string {
	host := serverURL.Hostname()
	if serverURL.Port() != "" {
		host += ":" + serverURL.Port()
	}
	return host
}

func GetToken() (string, error) {
	creds, err := readCreds()
	if err != nil {
		return "", err
	}

	tok, ok := creds[credsHost()]
	if !ok {
		tok = creds["default"]
	}

	return tok, nil
}

func StoreToken(token string) error {
	creds, err := readCreds()
	if err != nil {
		return err
	}

	host := credsHost()

	if token == "" {
		delete(creds, host)
	} else {
		creds[host] = token
	}

	var bs []byte

	// If creds are empty, make the file empty instead of yaml's '{}'.
	if len(creds) != 0 {
		if bs, err = yaml.Marshal(creds); err != nil {
			return err
		}
	}

	return os.WriteFile(credsPath, bs, 0o600)
}
