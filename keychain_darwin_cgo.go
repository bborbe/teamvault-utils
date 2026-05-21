//go:build darwin && cgo

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

/*
#include <Security/Security.h>
*/
import "C"

import (
	"context"
	"unsafe"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

func validatePasswordForKeychain(ctx context.Context, password Password) error {
	for i, ch := range password {
		if ch == 0 {
			return errors.Errorf(
				ctx,
				"password contains NUL byte at position %d which is not supported by Keychain",
				i,
			)
		}
		if ch == '\n' {
			return errors.Errorf(
				ctx,
				"password contains newline which is not supported by Keychain",
			)
		}
	}
	return nil
}

func (d *darwinKeychain) WritePassword(ctx context.Context, url Url, password Password) error {
	if err := validatePasswordForKeychain(ctx, password); err != nil {
		return err
	}
	if url == "" {
		glog.V(3).Infof("keychain write skipped: empty URL")
		return nil
	}

	serviceName := C.CString(KeychainServiceName)
	defer C.free(unsafe.Pointer(serviceName))

	accountName := C.CString(string(url))
	defer C.free(unsafe.Pointer(accountName))

	passwordBytes := []byte(password)
	passwordPtr := (*C.uchar)(C.CBytes(passwordBytes))
	passwordLen := C.uint32_t(len(passwordBytes))
	defer C.free(unsafe.Pointer(passwordPtr))

	var itemRef C.SecKeychainItemRef
	status := C.SecKeychainAddGenericPassword(
		nil,
		C.uint32_t(len(KeychainServiceName)),
		serviceName,
		C.uint32_t(len(string(url))),
		accountName,
		passwordLen,
		unsafe.Pointer(passwordPtr),
		&itemRef,
	)

	if status == C.errSecDuplicateItem {
		glog.V(4).Infof("keychain item already exists for url %q, updating", url)
		var existingItem C.SecKeychainItemRef
		findStatus := C.SecKeychainFindGenericPassword(
			nil,
			C.uint32_t(len(KeychainServiceName)),
			serviceName,
			C.uint32_t(len(string(url))),
			accountName,
			nil,
			nil,
			&existingItem,
		)
		if findStatus != C.errSecSuccess {
			return errors.Errorf(
				ctx,
				"security find-generic-password failed to find existing item: %d",
				findStatus,
			)
		}
		modifyStatus := C.SecKeychainItemModifyAttributesAndData(
			existingItem,
			0,
			nil,
			passwordLen,
			unsafe.Pointer(passwordPtr),
		)
		if modifyStatus != C.errSecSuccess {
			return errors.Errorf(
				ctx,
				"security SecKeychainItemModifyAttributesAndData failed with status %d",
				modifyStatus,
			)
		}
	} else if status != C.errSecSuccess {
		if status == C.errSecAuthFailed || status == C.errSecInteractionNotAllowed {
			glog.V(2).Infof("keychain locked for url %q", url)
			return errors.Errorf(
				ctx,
				"TeamVault password requires Keychain unlock; unlock your Keychain and retry",
			)
		}
		return errors.Errorf(ctx, "security SecKeychainAddGenericPassword failed with status %d", status)
	}

	glog.V(2).Infof("keychain write succeeded for url %q", url)
	return nil
}
