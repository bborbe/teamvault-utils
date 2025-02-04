// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bborbe/errors"
)

type RequestFailedError struct {
	Method     string
	URL        string
	StatusCode int
}

func (r RequestFailedError) Error() string {
	return fmt.Sprintf("%s request to %s failed with statusCode %d", r.Method, r.URL, r.StatusCode)
}

var NotFound = stderrors.New("not found")

func addRequestResponseToError(err error, resp *http.Response, req *http.Request) error {
	data := make(map[string]string)
	if req != nil {
		data["method"] = req.Method
		data["url"] = req.URL.String()
	}
	if resp != nil {
		data["status_code"] = strconv.Itoa(resp.StatusCode)
		data["status"] = resp.Status
	}
	return errors.AddDataToError(
		err,
		data,
	)
}

func CheckResponseIsSuccessful(req *http.Request, resp *http.Response) error {
	if resp.StatusCode == 404 {
		return errors.Wrapf(
			req.Context(),
			NotFound,
			"%s to %s failed with status %d",
			req.Method,
			req.URL.String(),
			resp.StatusCode,
		)
	}
	if resp.StatusCode/100 != 2 && resp.StatusCode/100 != 3 {
		content, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewBuffer(content))
		return errors.AddDataToError(
			errors.Wrapf(
				req.Context(),
				RequestFailedError{
					Method:     req.Method,
					URL:        req.URL.String(),
					StatusCode: resp.StatusCode,
				},
				"request failed content: %s",
				string(content),
			),
			map[string]string{
				"status_code": strconv.Itoa(resp.StatusCode),
				"status":      resp.Status,
				"method":      req.Method,
				"url":         req.URL.String(),
				"body":        string(content),
			},
		)
	}
	return nil
}
