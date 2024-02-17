// Adapted from https://github.com/qri-io/starlib/blob/master/http/http.go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2/clientcredentials"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// TODO(ENG-242): limit outreach ("RequestGuard" at https://github.com/qri-io/starlib/blob/master/http/http.go#L59)

// Encodings for form data.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
const (
	contentTypeMultipart = "multipart/form-data"
	contentTypeForm      = "application/x-www-form-urlencoded"

	jsonContentType = "application/json"
)

// request is a factory function for generating autokitteh
// functions for different HTTP request methods.
func request(method string) sdkexecutor.Function {
	return func(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		// Parse the input arguments.
		var (
			rawURL, contentType, body string
			headers, params, formData map[string]string
			jsonBody                  sdktypes.Value
			basicAuth                 [2]string
			oauth2Config              *clientcredentials.Config
		)

		err := sdkmodule.UnpackArgs(args, kwargs,
			"url", &rawURL,
			"params?", &params,
			"headers?", &headers,
			"body?", &body,
			"form_data?", &formData,
			// TODO: Content-Type is a header, also what about JSON? Charset?
			"content_type?", &contentType,
			"json_body?", &jsonBody,
			"basic_auth=?", &basicAuth,
			"oauth2=?", &oauth2Config,
		)
		if err != nil {
			return nil, err
		}

		if err := setQueryParams(&rawURL, params); err != nil {
			return nil, err
		}

		// Construct and send HTTP request.
		req, err := http.NewRequest(method, rawURL, nil)
		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			req.Header.Add(k, v)
		}

		if basicAuth[0] != "" || basicAuth[1] != "" {
			req.SetBasicAuth(basicAuth[0], basicAuth[1])
		}

		httpClient := http.DefaultClient

		if oauth2Config != nil {
			httpClient = oauth2Config.Client(ctx)
		}

		if err = setBody(req, body, formData, contentType, jsonBody); err != nil {
			return nil, err
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Parse and return the response.
		return toStruct(ctx, res)
	}
}

func setQueryParams(rawurl *string, params map[string]string) error {
	u, err := url.Parse(*rawurl)
	if err != nil {
		return err
	}

	q := u.Query()

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	*rawurl = u.String()
	return nil
}

func setBody(req *http.Request, rawBody string, formData map[string]string, contentType string, jsonData sdktypes.Value) error {
	errMutuallyExclusive := errors.New("body, form_data and json_data are mutually exclusive")

	if rawBody != "" {
		if (jsonData != nil && !sdktypes.IsNothingValue(jsonData)) || formData != nil {
			return errMutuallyExclusive
		}

		req.Body = io.NopCloser(strings.NewReader(rawBody))

		// Specifying the Content-Length ensures that https://go.dev/src/net/http/transfer.go
		// doesnt specify Transfer-Encoding: chunked which is not supported by some endpoints.
		// This is required when using ioutil.NopCloser method for the request body
		// (see ShouldSendChunkedRequestBody() in the library mentioned above).
		req.ContentLength = int64(len(rawBody))

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		return nil
	}

	if jsonData != nil && !sdktypes.IsNothingValue(jsonData) {
		if formData != nil {
			return errMutuallyExclusive
		}

		if contentType == "" {
			contentType = jsonContentType
		}

		req.Header.Set("Content-Type", contentType)

		v, err := sdkvalues.ValueWrapper{SafeForJSON: true}.Unwrap(jsonData)
		if err != nil {
			return err
		}

		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		req.Body = io.NopCloser(bytes.NewBuffer(data))
		req.ContentLength = int64(len(data))

		return nil
	}

	if formData != nil {
		var form url.Values

		for k, v := range formData {
			form.Add(k, v)
		}

		switch strings.Split(contentType, ";")[0] {
		case contentTypeForm, "":
			contentType = contentTypeForm
			req.Body = io.NopCloser(strings.NewReader(form.Encode()))
			req.ContentLength = int64(len(form.Encode()))

		case contentTypeMultipart:
			var b bytes.Buffer
			mw := multipart.NewWriter(&b)
			defer mw.Close()

			contentType = mw.FormDataContentType()

			for k, values := range form {
				for _, v := range values {
					w, err := mw.CreateFormField(k)
					if err != nil {
						return err
					}
					if _, err := w.Write([]byte(v)); err != nil {
						return err
					}
				}
			}

			req.Body = io.NopCloser(&b)

		default:
			return fmt.Errorf("unknown form encoding: %s", contentType)
		}

		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	return nil
}

func toStruct(ctx context.Context, r *http.Response) (sdktypes.Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	// reset reader to allow multiple calls
	r.Body = io.NopCloser(bytes.NewReader(body))

	var (
		jsonValue sdktypes.Value
		data      any
	)

	if err := json.Unmarshal(body, &data); err == nil {
		jsonValue, _ = sdkvalues.Wrap(data)
	}

	if jsonValue == nil {
		jsonValue = sdktypes.NewNothingValue()
	}

	return sdktypes.NewStructValue(
		sdktypes.NewStringValue("http_response"),
		map[string]sdktypes.Value{
			"url":         sdktypes.NewStringValue(r.Request.URL.String()),
			"status_code": sdktypes.NewIntegerValue(int64(r.StatusCode)),
			"headers": sdktypes.NewDictValue(
				kittehs.TransformMapToList(r.Header, func(k string, vs []string) *sdktypes.DictValueItem {
					return &sdktypes.DictValueItem{
						K: sdktypes.NewStringValue(k),
						V: sdktypes.NewStringValue(strings.Join(vs, ",")),
					}
				}),
			),
			"encoding":    sdktypes.NewStringValue(strings.Join(r.TransferEncoding, ",")),
			"body_bytes":  sdktypes.NewBytesValue(body),
			"body_json":   jsonValue,
			"body_string": sdktypes.NewStringValue(string(body)),
			"body":        sdktypes.NewStringValue(string(body)),
		}), nil
}
