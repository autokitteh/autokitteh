// Adapted from https://github.com/qri-io/starlib/blob/master/http/http.go.

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

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(ENG-242): limit outreach ("RequestGuard" at
// https://github.com/qri-io/starlib/blob/master/http/http.go#L59)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("http"))

const (
	authHeader        = "Authorization"
	contentTypeHeader = "Content-Type"

	contentTypeForm      = "application/x-www-form-urlencoded"
	contentTypeJSON      = "application/json"
	contentTypeMultipart = "multipart/form-data"
)

const (
	bodyTypeRaw  = "raw"
	bodyTypeJSON = "json"
	bodyTypeForm = "form"
)

type request struct {
	url             string
	headers, params map[string]string
	body            *bytes.Buffer
	bodyType        string
	contentLen      int64
}

func New() sdkexecutor.Executor {
	args := sdkmodule.WithArgs(
		"url",
		"params?",
		"headers?",
		"data?",
		"json?",
	)

	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportFunction(
			"delete",
			doRequest(http.MethodDelete),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#DELETE"),
			args),
		sdkmodule.ExportFunction(
			"get",
			doRequest(http.MethodGet),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#GET"),
			args),
		sdkmodule.ExportFunction(
			"head",
			doRequest(http.MethodHead),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#HEAD"),
			args),
		sdkmodule.ExportFunction(
			"options",
			doRequest(http.MethodOptions),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#OPTIONS"),
			args),
		sdkmodule.ExportFunction(
			"patch",
			doRequest(http.MethodPatch),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc5789"),
			args),
		sdkmodule.ExportFunction(
			"post",
			doRequest(http.MethodPost),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#POST"),
			args),
		sdkmodule.ExportFunction(
			"put",
			doRequest(http.MethodPut),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#PUT"),
			args),
	)
}

// doRequest is a factory function for generating autokitteh
// functions for different HTTP request methods.
func doRequest(method string) sdkexecutor.Function {
	return func(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		// Parse the input arguments.
		var (
			err error
			req request
		)

		// parse args and kwargs
		if err = unpackAndParseArgs(&req, method, args, kwargs); err != nil {
			return sdktypes.InvalidValue, err
		}

		// Construct and send HTTP request.
		res, err := sendHttpRequest(ctx, req, method)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		// Parse and return the response.
		return toStruct(res)
	}
}

func unpackAndParseArgs(req *request, method string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (err error) {
	var data sdktypes.Value

	if len(args) > 1 { // just to have a better error message instead of "not all args consumed"
		return errors.New("pass non-URL arguments as kwargs only")
	}

	if err = sdkmodule.UnpackArgs(args, kwargs,
		"url", &req.url,
		"params=?", &req.params,
		"headers=?", &req.headers,
		"json=?", &data, // alias for data
		"data=?", &data, // will override json, if both json and data are provided
	); err != nil {
		return err
	}
	if req.headers == nil {
		req.headers = make(map[string]string)
	}

	// NOTE: GET request shouldn't have user-defined body.
	// Python's requests lib will ignore body on GET as well
	if method != http.MethodGet && data.IsValid() {

		// if data passed as JSON and data is not passed, then request to parse as JSON content
		if _, isJson := kwargs["json"]; isJson {
			if _, isData := kwargs["data"]; !isData {
				req.headers[contentTypeHeader] = contentTypeJSON
			}
		}

		if err = parseBody(req, data); err != nil {
			return err
		}
	}

	if err := setQueryParams(&req.url, req.params); err != nil {
		return err
	}
	return nil
}

// parses provided body and updates headers accordingly
func parseBody(req *request, body sdktypes.Value) (err error) {
	var (
		rawBody, contentType string
		formBody             map[string]string
		jsonBody             sdktypes.Value
		ok                   bool
	)

	if contentType, ok = req.headers[contentTypeHeader]; ok { // use content type, if provided
		switch contentType {
		case contentTypeJSON:
			req.bodyType = bodyTypeJSON
		case contentTypeForm, contentTypeMultipart:
			req.bodyType = bodyTypeForm
		}
	}

	// parse bodyType. RAW -> FORM -> JSON, unless specific type is requested
	if req.bodyType == "" {
		if err = body.UnwrapInto(&rawBody); err == nil {
			req.bodyType = bodyTypeRaw
		}
	}
	if (err != nil && req.bodyType == "") || req.bodyType == bodyTypeForm {
		if err = body.UnwrapInto(&formBody); err == nil {
			req.bodyType = bodyTypeForm
			if contentType == "" {
				req.headers[contentTypeHeader] = contentTypeForm
			}
		}
	}
	if (err != nil && req.bodyType == "") || req.bodyType == bodyTypeJSON {
		if err = body.UnwrapInto(&jsonBody); err == nil {
			req.bodyType = bodyTypeJSON
			if contentType == "" {
				req.headers[contentTypeHeader] = contentTypeJSON
			}
		}
	}

	if err != nil {
		return errors.New("body must be one of <string|form|json>")
	}

	// parse body
	switch req.bodyType {
	case bodyTypeRaw:
		req.body = bytes.NewBufferString(rawBody)

		// Specifying the Content-Length ensures that https://go.dev/src/net/http/transfer.go
		// doesnt specify Transfer-Encoding: chunked which is not supported by some endpoints.
		// This is required when using ioutil.NopCloser method for the request body
		// (see ShouldSendChunkedRequestBody() in the library mentioned above).
		req.contentLen = int64(len(rawBody))

	case bodyTypeJSON:
		if !jsonBody.IsValid() || jsonBody.IsNothing() {
			return nil
		}
		v, err := sdktypes.ValueWrapper{SafeForJSON: true}.Unwrap(jsonBody)
		if err != nil {
			return err
		}
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		req.body = bytes.NewBuffer(data)
		req.contentLen = int64(len(data))

	case bodyTypeForm:
		if formBody == nil {
			return nil
		}
		form := make(url.Values)
		for k, v := range formBody {
			form.Add(k, v)
		}

		// Ignore (but allow the user to set) the charset in the Content-Type header.
		switch strings.Split(req.headers[contentType], ";")[0] {
		case "", contentTypeForm:
			s := form.Encode()
			req.body = bytes.NewBufferString(s)
			req.contentLen = int64(len(s))

		case contentTypeMultipart:
			mw := multipart.NewWriter(req.body)
			defer mw.Close()
			req.headers[contentTypeHeader] = mw.FormDataContentType()

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
			// TODO: should we set the contentLen?

		default:
			return fmt.Errorf("unknown form encoding: %s", contentType)
		}

	}

	return nil
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

func createHttpRequest(ctx context.Context, req request, method string) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, req.url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range req.headers {
		httpReq.Header.Set(k, v)
	}

	if req.contentLen != 0 {
		httpReq.ContentLength = req.contentLen
	}

	if req.body != nil {
		httpReq.Body = io.NopCloser(req.body)
	}
	return httpReq, nil
}

// construct and send HTTP request
func sendHttpRequest(ctx context.Context, req request, method string) (*http.Response, error) {
	httpReq, err := createHttpRequest(ctx, req, method)
	if err != nil {
		return nil, err
	}

	httpClient := http.DefaultClient

	res, err := httpClient.Do(httpReq)
	if err != nil {
		if uerr := new(url.Error); errors.As(err, &uerr) {
			err = sdktypes.NewProgramError(
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
		return nil, err
	}
	return res, nil
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
			"status_code": sdktypes.NewIntegerValue(r.StatusCode),
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

	// add form() only for requests (when not nil) and not for responses
	if form != nil {
		formBody := kittehs.Must1(sdktypes.NewConstFunctionValue("form", sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(form, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ","))
			}),
		)))
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
	return kittehs.Must1(sdktypes.NewStructValue(
		sdktypes.NewStringValue("body"),
		map[string]sdktypes.Value{
			"text":  kittehs.Must1(sdktypes.NewConstFunctionValue("text", sdktypes.NewStringValue(string(body)))),
			"bytes": kittehs.Must1(sdktypes.NewConstFunctionValue("bytes", sdktypes.NewBytesValue(body))),
			"json":  jsonBody,
		},
	))
}
