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

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(ENG-242): limit outreach ("RequestGuard" at
// https://github.com/qri-io/starlib/blob/master/http/http.go#L59)

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"

	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeJSON      = "application/json"
	contentTypeMultipart = "multipart/form-data"
)

var args = sdkmodule.WithArgs(
	"url",
	"params?",
	"headers?",
	"raw_body?",
	"form_body?",
	"json_body?",
)

// request is a factory function for generating autokitteh
// functions for different HTTP request methods.
func (i integration) request(method string) sdkexecutor.Function {
	return func(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		// Parse the input arguments.
		var (
			rawURL, rawBody           string
			headers, params, formBody map[string]string
			jsonBody                  sdktypes.Value
		)
		err := sdkmodule.UnpackArgs(args, kwargs,
			"url", &rawURL,
			"params?", &params,
			"headers?", &headers,
			"raw_body?", &rawBody,
			"form_body?", &formBody,
			"json_body?", &jsonBody,
		)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if err := setQueryParams(&rawURL, params); err != nil {
			return sdktypes.InvalidValue, err
		}

		// Add the Authorization HTTP header?
		if auth := i.getConnection(ctx)["authorization"]; auth != "" {
			// If the Authorization header is set explicitly, it
			// should override the connection's default authorization.
			if _, ok := headers[authHeader]; !ok {
				headers[authHeader] = auth
			}
		}

		// Construct and send HTTP request.
		req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		if err = setBody(req, rawBody, formBody, jsonBody); err != nil {
			return sdktypes.InvalidValue, err
		}

		httpClient := http.DefaultClient

		res, err := httpClient.Do(req)
		if err != nil {
			if uerr := new(url.Error); errors.As(err, &uerr) {
				return sdktypes.InvalidValue, sdktypes.NewProgramError(
					kittehs.Must1(sdktypes.NewStructValue(
						sdktypes.NewStringValue("url_error"),
						map[string]sdktypes.Value{
							"url":       sdktypes.NewStringValue(uerr.URL),
							"op":        sdktypes.NewStringValue(uerr.Op),
							"temporary": sdktypes.NewBooleanValue(uerr.Temporary()),
							"timeout":   sdktypes.NewBooleanValue(uerr.Timeout()),
							"error":     sdktypes.NewStringValue(uerr.Err.Error()),
						},
					)),
					nil,
					nil,
				).ToError()
			}

			return sdktypes.InvalidValue, err
		}

		// Parse and return the response.
		return toStruct(res)
	}
}

// setQueryParams updates the given URL, based on the given query parameters.
func setQueryParams(rawURL *string, params map[string]string) error {
	u, err := url.Parse(*rawURL)
	if err != nil {
		return err
	}

	q := u.Query()

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	*rawURL = u.String()
	return nil
}

// getConnection returns the secret data associated with this connection, if there is any.
func (i integration) getConnection(ctx context.Context) map[string]string {
	// Extract the connection token from the given context.
	cfg := sdkmodule.FunctionDataFromContext(ctx)
	if cfg == nil {
		cfg = []byte{}
	}

	c, err := i.secrets.Get(ctx, i.scope, string(cfg))
	if err != nil {
		return nil
	}
	return c
}

func setBody(req *http.Request, rawBody string, formBody map[string]string, jsonBody sdktypes.Value) error {
	errMutuallyExclusive := errors.New("raw_body, form_body, and json_body are mutually exclusive")

	// Raw body with unknown content type.
	if rawBody != "" {
		if formBody != nil || (jsonBody.IsValid() && !jsonBody.IsNothing()) {
			return errMutuallyExclusive
		}

		req.Body = io.NopCloser(strings.NewReader(rawBody))

		// Specifying the Content-Length ensures that https://go.dev/src/net/http/transfer.go
		// doesnt specify Transfer-Encoding: chunked which is not supported by some endpoints.
		// This is required when using ioutil.NopCloser method for the request body
		// (see ShouldSendChunkedRequestBody() in the library mentioned above).
		req.ContentLength = int64(len(rawBody))
		return nil
	}

	// JSON body.
	if jsonBody.IsValid() && !jsonBody.IsNothing() {
		if formBody != nil {
			return errMutuallyExclusive
		}

		// Set the Content-Type header only if it's not already set.
		if req.Header.Get(contentTypeHeader) == "" {
			req.Header.Set(contentTypeHeader, contentTypeJSON)
		}

		v, err := sdktypes.ValueWrapper{SafeForJSON: true}.Unwrap(jsonBody)
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

	if formBody != nil {
		form := make(url.Values)
		for k, v := range formBody {
			form.Add(k, v)
		}

		// Set the Content-Type header only if it's not already set.
		contentType := req.Header.Get(contentTypeHeader)
		if contentType == "" {
			req.Header.Set(contentTypeHeader, contentTypeForm)
		}

		// Ignore (but allow the user to set) the charset in the Content-Type header.
		switch strings.Split(contentType, ";")[0] {
		case contentTypeForm:
			s := form.Encode()
			req.Body = io.NopCloser(strings.NewReader(s))
			req.ContentLength = int64(len(s))

		case contentTypeMultipart:
			var b bytes.Buffer
			mw := multipart.NewWriter(&b)
			defer mw.Close()

			req.Header.Set(contentTypeHeader, mw.FormDataContentType())

			for k, vs := range form {
				for _, v := range vs {
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
	}

	return nil
}

// toStruct converts an HTTP response to an autokitteh struct.
func toStruct(r *http.Response) (sdktypes.Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	r.Body.Close()

	// Reset reader to allow multiple calls.
	r.Body = io.NopCloser(bytes.NewReader(body))

	return sdktypes.NewStructValue(
		sdktypes.NewStringValue("http_response"),
		map[string]sdktypes.Value{
			"url":         sdktypes.NewStringValue(r.Request.URL.String()),
			"status_code": sdktypes.NewIntegerValue(int64(r.StatusCode)),
			"headers": kittehs.Must1(sdktypes.NewDictValue(
				kittehs.TransformMapToList(r.Header, func(k string, vs []string) sdktypes.DictItem {
					return sdktypes.DictItem{
						K: sdktypes.NewStringValue(k),
						V: sdktypes.NewStringValue(strings.Join(vs, ",")),
					}
				}),
			)),
			"encoding": sdktypes.NewStringValue(strings.Join(r.TransferEncoding, ",")),
			"body":     bodyToStruct(body, nil),
		})
}

func bodyToStruct(body []byte, form url.Values) sdktypes.Value {
	var (
		v        any
		formBody sdktypes.Value = sdktypes.Nothing
		jsonBody sdktypes.Value
	)

	jsonDecoder := json.NewDecoder(bytes.NewReader(body))
	jsonDecoder.UseNumber()

	if err := jsonDecoder.Decode(&v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else if vv, err := sdktypes.WrapValue(v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionValue("json", vv))
	}

	// Right now, form is always nil. We need to handle this case.
	if form != nil {
		formBody = kittehs.Must1(sdktypes.NewConstFunctionValue("form", sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(form, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ","))
			}),
		)))
	}

	return kittehs.Must1(sdktypes.NewStructValue(
		sdktypes.NewStringValue("body"),
		map[string]sdktypes.Value{
			"text":  kittehs.Must1(sdktypes.NewConstFunctionValue("text", sdktypes.NewStringValue(string(body)))),
			"bytes": kittehs.Must1(sdktypes.NewConstFunctionValue("bytes", sdktypes.NewBytesValue(body))),
			"json":  jsonBody,
			"form":  formBody,
		},
	))
}
