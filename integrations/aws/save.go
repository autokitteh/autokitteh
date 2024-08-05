package aws

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger *zap.Logger
}

func NewHTTPHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}
	// Trim whitespace and retrieve form values
	region := strings.TrimSpace(r.FormValue("region"))
	accessKey := strings.TrimSpace(r.FormValue("access_key"))
	secretKey := strings.TrimSpace(r.FormValue("secret_key"))
	token := strings.TrimSpace(r.FormValue("token"))

	// Test the AWS authentication details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	api := AWSAPI{
		Vars: AWSVars{
			AccessKeyID:  accessKey,
			Region:       region,
			SecretKey:    secretKey,
			SessionToken: token,
		},
	}
	_, err := api.Test(ctx)
	if err != nil {
		l.Warn("AWS authentication test failed", zap.Error(err))
		c.AbortBadRequest("AWS authentication test failed: " + err.Error())
		return
	}

	c.Finalize(sdktypes.EncodeVars(authData{
		Region:      region,
		AccessKeyID: accessKey,
		SecretKey:   secretKey,
		Token:       token,
	}))
}
