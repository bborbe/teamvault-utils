// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"encoding/json"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

type StoreMapperTx[KEY ~[]byte | ~string, OBJECT any] interface {
	Map(ctx context.Context, tx Tx, fn func(ctx context.Context, key KEY, object OBJECT) error) error
}

type StoreAdderTx[KEY ~[]byte | ~string, OBJECT any] interface {
	Add(ctx context.Context, tx Tx, key KEY, object OBJECT) error
}

type StoreRemoverTx[KEY ~[]byte | ~string] interface {
	Remove(ctx context.Context, tx Tx, key KEY) error
}

type StoreGetterTx[KEY ~[]byte | ~string, OBJECT any] interface {
	Get(ctx context.Context, tx Tx, key KEY) (*OBJECT, error)
}

type StoreExistsTx[KEY ~[]byte | ~string, OBJECT any] interface {
	Exists(ctx context.Context, tx Tx, key KEY) (bool, error)
}

type StoreStreamTx[KEY ~[]byte | ~string, OBJECT any] interface {
	Stream(ctx context.Context, tx Tx, ch chan<- OBJECT) error
}

type StoreTx[KEY ~[]byte | ~string, OBJECT any] interface {
	StoreAdderTx[KEY, OBJECT]
	StoreRemoverTx[KEY]
	StoreGetterTx[KEY, OBJECT]
	StoreMapperTx[KEY, OBJECT]
	StoreExistsTx[KEY, OBJECT]
	StoreStreamTx[KEY, OBJECT]
}

func NewStoreTx[KEY ~[]byte | ~string, OBJECT any](bucketName BucketName) StoreTx[KEY, OBJECT] {
	return &storeTx[KEY, OBJECT]{
		bucketName: bucketName,
	}
}

type storeTx[KEY ~[]byte | ~string, OBJECT any] struct {
	bucketName BucketName
}

func (s storeTx[KEY, OBJECT]) Add(ctx context.Context, tx Tx, key KEY, object OBJECT) error {
	bucket, err := tx.CreateBucketIfNotExists(ctx, s.bucketName)
	if err != nil {
		return errors.Wrapf(ctx, err, "get bucket failed")
	}
	value, err := json.Marshal(object)
	if err != nil {
		return errors.Wrapf(ctx, err, "marshal json failed")
	}
	if err = bucket.Put(ctx, []byte(key), value); err != nil {
		return errors.Wrapf(ctx, err, "set failed")
	}
	return nil
}

func (s storeTx[KEY, OBJECT]) Remove(ctx context.Context, tx Tx, key KEY) error {
	bucket, err := tx.CreateBucketIfNotExists(ctx, s.bucketName)
	if err != nil {
		if errors.Is(err, BucketNotFoundError) {
			glog.V(3).Infof("bucket %s not found", s.bucketName)
			return nil
		}
		return errors.Wrapf(ctx, err, "get bucket failed")
	}
	if err := bucket.Delete(ctx, []byte(key)); err != nil {
		return errors.Wrapf(ctx, err, "remove %s failed", key)
	}
	return nil
}

func (s storeTx[KEY, OBJECT]) Get(ctx context.Context, tx Tx, key KEY) (*OBJECT, error) {
	var object OBJECT
	bucket, err := tx.Bucket(ctx, s.bucketName)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get bucket failed")
	}
	item, err := bucket.Get(ctx, []byte(key))
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "get %s failed", string(key))
	}
	err = item.Value(func(val []byte) error {
		if len(val) == 0 {
			return errors.Wrapf(ctx, KeyNotFoundError, "key(%s) not found", string(key))
		}
		return json.Unmarshal(val, &object)
	})
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "handel value failed")
	}
	return &object, nil
}

func (s storeTx[KEY, OBJECT]) Exists(ctx context.Context, tx Tx, key KEY) (bool, error) {
	bucket, err := tx.Bucket(ctx, s.bucketName)
	if err != nil {
		if errors.Is(err, BucketNotFoundError) {
			glog.V(3).Infof("bucket %s not found", s.bucketName)
			return false, nil
		}
		return false, errors.Wrapf(ctx, err, "get bucket failed")
	}
	item, err := bucket.Get(ctx, []byte(key))
	if err != nil {
		return false, errors.Wrapf(ctx, err, "get %s failed", string(key))
	}
	return item.Exists(), nil
}

func (s storeTx[KEY, OBJECT]) Map(ctx context.Context, tx Tx, fn func(ctx context.Context, key KEY, object OBJECT) error) error {
	bucket, err := tx.Bucket(ctx, s.bucketName)
	if err != nil {
		if errors.Is(err, BucketNotFoundError) {
			glog.V(3).Infof("bucket %s not found", s.bucketName)
			return nil
		}
		return errors.Wrapf(ctx, err, "get bucket failed")
	}
	it := bucket.Iterator()
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			item := it.Item()
			key := KEY(item.Key())
			err := item.Value(func(v []byte) error {
				var object OBJECT
				if err := json.Unmarshal(v, &object); err != nil {
					return errors.Wrapf(ctx, err, "unmarshal %s failed", string(key))
				}
				if err := fn(ctx, key, object); err != nil {
					return errors.Wrapf(ctx, err, "call fn failed")
				}
				return nil
			})
			if err != nil {
				return errors.Wrapf(ctx, err, "handle value failed")
			}
		}
	}
	return nil
}

func (s storeTx[KEY, OBJECT]) Stream(ctx context.Context, tx Tx, ch chan<- OBJECT) error {
	return s.Map(ctx, tx, func(ctx context.Context, key KEY, object OBJECT) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- object:
			return nil
		}
	})
}
