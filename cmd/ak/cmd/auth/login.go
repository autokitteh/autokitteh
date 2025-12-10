package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var loginCmd = common.StandardCommand(&cobra.Command{
	Use:   "login",
	Short: "Login",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, _ []string) error {
		port, wait, done, err := initServer()
		defer done()

		if err != nil {
			return err
		}

		link := fmt.Sprintf("%s/auth/cli-login?p=%d", common.ServerURL(), port)
		if err := browser.OpenURL(link); err != nil {
			return err
		}

		token, err := wait(cmd.Context())
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.ErrOrStderr(), "You are now logged in.")

		return common.StoreToken(token)
	},
})

func initServer() (port int, wait func(context.Context) (string, error), done func(), err error) {
	type resp struct {
		token string
		err   error
	}

	ch := make(chan *resp)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ch <- &resp{token: r.URL.Query().Get("token")}
		fmt.Fprintf(w, "You can close this tab now.")
	})

	var l net.Listener
	if l, err = net.Listen("tcp", ":0"); err != nil {
		return
	}

	go func() {
		if err = http.Serve(l, nil); err != nil {
			ch <- &resp{err: err}
		}
	}()

	done = func() { _ = l.Close() }

	port = l.Addr().(*net.TCPAddr).Port

	wait = func(ctx context.Context) (string, error) {
		select {
		case resp := <-ch:
			return resp.token, resp.err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return
}
