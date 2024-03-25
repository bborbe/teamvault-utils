// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

type ErrorList interface {
	Add(err error)
	Len() int
	AsStringArray() []string
}

type errorList struct {
	errs []error
}

func NewErrorList() ErrorList {
	return &errorList{}
}

func (e *errorList) Add(err error) {
	e.errs = append(e.errs, err)
}

func (e *errorList) Len() int {
	return len(e.errs)
}

func (e *errorList) AsStringArray() []string {
	var errs []string
	if len(e.errs) > 0 {
		for _, invoiceErr := range e.errs {
			errs = append(errs, invoiceErr.Error())
		}
	}
	return errs
}
