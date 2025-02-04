// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func BasicTestSuite(provider Provider) {
	GinkgoHelper()
	Context("Basic", func() {
		var db DB
		var err error
		var bucketName BucketName
		var ctx context.Context
		BeforeEach(func() {
			ctx = context.Background()
			db, err = provider.Get(ctx)
			Expect(db).NotTo(BeNil())
			Expect(err).To(BeNil())
			bucketName = NewBucketName("mybucket")
		})
		It("Get not found", func() {
			err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
				bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
				Expect(err).To(BeNil())

				{
					item, err := bucket.Get(ctx, []byte("key"))
					Expect(err).To(BeNil())
					Expect(item.Key()).To(Equal([]byte("key")))
					var result []byte
					err = item.Value(func(val []byte) error {
						result = val
						return nil
					})
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())
				}

				return nil
			})
			Expect(err).To(BeNil())
		})
		It("Put and Get", func() {
			err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
				bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
				Expect(err).To(BeNil())

				{
					err = bucket.Put(ctx, []byte("key"), []byte("value"))
					Expect(err).To(BeNil())
				}

				{
					item, err := bucket.Get(ctx, []byte("key"))
					Expect(err).To(BeNil())
					Expect(item.Key()).To(Equal([]byte("key")))
					var result []byte
					err = item.Value(func(val []byte) error {
						result = val
						return nil
					})
					Expect(err).To(BeNil())
					Expect(result).To(Equal([]byte("value")))
				}

				return nil
			})
			Expect(err).To(BeNil())
		})
		It("Put, Delete and Get", func() {
			err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
				bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
				Expect(err).To(BeNil())

				{
					err = bucket.Put(ctx, []byte("key"), []byte("value"))
					Expect(err).To(BeNil())
				}

				{
					err = bucket.Delete(ctx, []byte("key"))
					Expect(err).To(BeNil())
				}

				{
					item, err := bucket.Get(ctx, []byte("key"))
					Expect(err).To(BeNil())
					Expect(item.Key()).To(Equal([]byte("key")))
					var result []byte
					err = item.Value(func(val []byte) error {
						result = val
						return nil
					})
					Expect(err).To(BeNil())
					Expect(result).To(BeNil())
				}

				return nil
			})
			Expect(err).To(BeNil())
		})
		It("tx in tx return error", func() {
			err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
				return db.Update(ctx, func(ctx context.Context, tx Tx) error {
					return nil
				})
			})
			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, TransactionAlreadyOpenError)).To(BeTrue())
		})
	})
}
