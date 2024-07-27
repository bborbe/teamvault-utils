// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentry

import (
	"context"
	"fmt"
	"io"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/getsentry/sentry-go"
	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/sentry-client.go --fake-name SentryClient . Client
type Client interface {
	CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	CaptureException(exception error, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	Flush(timeout stdtime.Duration) bool
	io.Closer
}

func NewClient(ctx context.Context, clientOptions sentry.ClientOptions, excludeErrors ...ExcludeError) (Client, error) {
	newClient, err := sentry.NewClient(clientOptions)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create sentry client failed")
	}
	newClient.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		if hint.Context != nil {
			for k, v := range errors.DataFromContext(hint.Context) {
				event.Tags[k] = v
			}
		}
		if hint.OriginalException != nil {
			for k, v := range errors.DataFromError(hint.OriginalException) {
				event.Tags[k] = v
			}
		}
		switch data := hint.Data.(type) {
		case map[string]interface{}:
			for k, v := range data {
				if v == nil {
					continue
				}
				event.Tags[k] = fmt.Sprintf("%v", v)
			}
		case map[string]string:
			for k, v := range data {
				event.Tags[k] = v
			}
		}
		return event
	})
	return &client{
		client:        newClient,
		excludeErrors: excludeErrors,
	}, nil
}

type client struct {
	client        *sentry.Client
	excludeErrors ExcludeErrors
}

func (c *client) Flush(timeout stdtime.Duration) bool {
	return c.client.Flush(timeout)
}

func (c *client) CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID {
	eventID := c.client.CaptureMessage(message, hint, scope)
	glog.V(2).Infof("capture sentry message with id %s: %s", *eventID, message)
	return eventID
}

func (c *client) CaptureException(err error, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID {
	if c.excludeErrors.IsExcluded(err) {
		glog.V(4).Infof("capture error %v is excluded => skip", err)
		return nil
	}
	if scope == nil {
		scope = sentry.NewScope()
	}
	if hint == nil {
		hint = &sentry.EventHint{}
	}
	if hint.OriginalException == nil {
		hint.OriginalException = err
	}
	eventID := c.client.CaptureException(err, hint, scope)
	glog.V(2).Infof("capture sentry execption with id %s: %v", *eventID, err)
	return eventID
}

func (c *client) Close() error {
	c.client.Flush(2 * stdtime.Second)
	return nil
}
