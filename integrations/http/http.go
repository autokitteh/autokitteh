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
)

// TODO(ENG-242): limit outreach ("RequestGuard" at https://github.com/qri-io/starlib/blob/master/http/http.go#L59)

// Encodings for form data.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/POST
const (
	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeJSON      = "application/json"
	contentTypeMultipart = "multipart/form-data"
)

// request is a factory function for generating autokitteh
// functions for different HTTP request methods.
func request(method string) sdkexecutor.Function {
	return func(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		// Parse the input arguments.
		var (
			rawURL, contentType, rawBody string
			headers, params, formBody    map[string]string
			jsonBody                     sdktypes.Value
			basicAuth                    [2]string
			oauth2Config                 *clientcredentials.Config
		)

		err := sdkmodule.UnpackArgs(args, kwargs,
			"url", &rawURL,
			"params?", &params,
			"headers?", &headers,
			"raw_body?", &rawBody,
			"form_body?", &formBody,
			"json_body?", &jsonBody,
			// TODO: Content-Type is a header, also what about JSON? Charset?
			"content_type?", &contentType,
			"basic_auth=?", &basicAuth,
			"oauth2=?", &oauth2Config,
		)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if err := setQueryParams(&rawURL, params); err != nil {
			return sdktypes.InvalidValue, err
		}

		// Construct and send HTTP request.
		req, err := http.NewRequest(method, rawURL, nil)
		if err != nil {
			return sdktypes.InvalidValue, err
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

		if err = setBody(req, rawBody, formBody, contentType, jsonBody); err != nil {
			return sdktypes.InvalidValue, err
		}

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

func setBody(req *http.Request, rawBody string, formBody map[string]string, contentType string, jsonBody sdktypes.Value) error {
	errMutuallyExclusive := errors.New("raw_body, form_body, and json_body are mutually exclusive")

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

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		return nil
	}

	if jsonBody.IsValid() && !jsonBody.IsNothing() {
		if formBody != nil {
			return errMutuallyExclusive
		}

		if contentType == "" {
			contentType = contentTypeJSON
		}

		req.Header.Set("Content-Type", contentType)

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

		switch strings.Split(contentType, ";")[0] {
		case "", contentTypeForm:
			contentType = contentTypeForm
			s := form.Encode()
			req.Body = io.NopCloser(strings.NewReader(s))
			req.ContentLength = int64(len(s))

		case contentTypeMultipart:
			var b bytes.Buffer
			mw := multipart.NewWriter(&b)
			defer mw.Close()

			contentType = mw.FormDataContentType()

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

		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	return nil
}

func toStruct(r *http.Response) (sdktypes.Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	r.Body.Close()

	// reset reader to allow multiple calls
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

	if err := json.Unmarshal(body, &v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else if vv, err := sdktypes.WrapValue(v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionValue("json", vv))
	}

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
