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

func BucketTestSuite(provider Provider) {
	GinkgoHelper()
	Context("Bucket", func() {
		var bucketName BucketName
		var db DB
		var err error
		var ctx context.Context
		BeforeEach(func() {
			ctx = context.Background()
			db, err = provider.Get(ctx)
			Expect(db).NotTo(BeNil())
			Expect(err).To(BeNil())
			bucketName = NewBucketName("mybucket")
		})
		Context("CreateBucket", func() {
			var bucket Bucket
			JustBeforeEach(func() {
				err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err = tx.CreateBucket(ctx, bucketName)
					return err
				})
			})
			Context("success", func() {
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns bucket", func() {
					Expect(bucket).NotTo(BeNil())
				})
			})
			Context("twise", func() {
				BeforeEach(func() {
					Expect(db.Update(ctx, func(ctx context.Context, tx Tx) error {
						_, err = tx.CreateBucket(ctx, bucketName)
						return err
					})).To(BeNil())
				})
				It("returns error", func() {
					Expect(err).NotTo(BeNil())
					Expect(errors.Is(err, BucketAlreadyExistsError)).To(BeTrue())
				})
				It("returns nil bucket", func() {
					Expect(bucket).To(BeNil())
				})
			})
		})
		Context("CreateBucketIfNotExists", func() {
			var bucket Bucket
			JustBeforeEach(func() {
				err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err = tx.CreateBucketIfNotExists(ctx, bucketName)
					return err
				})
			})
			Context("success", func() {
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns bucket", func() {
					Expect(bucket).NotTo(BeNil())
				})
			})
			Context("twise", func() {
				var secondBucket Bucket
				var thirdBucket Bucket
				BeforeEach(func() {
					Expect(db.Update(ctx, func(ctx context.Context, tx Tx) error {
						secondBucket, err = tx.CreateBucketIfNotExists(ctx, bucketName)
						if err != nil {
							return err
						}
						thirdBucket, err = tx.CreateBucketIfNotExists(ctx, bucketName)
						if err != nil {
							return err
						}
						return nil
					})).To(BeNil())
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns bucket", func() {
					Expect(bucket).NotTo(BeNil())
				})
				It("return same bucket in same tx", func() {
					Expect(secondBucket).NotTo(BeNil())
					Expect(thirdBucket).NotTo(BeNil())
					Expect(secondBucket).To(Equal(thirdBucket))
				})
			})
		})
		Context("Bucket", func() {
			var bucket Bucket
			JustBeforeEach(func() {
				err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err = tx.Bucket(ctx, bucketName)
					return err
				})
			})
			Context("success", func() {
				var secondBucket Bucket
				var thirdBucket Bucket
				BeforeEach(func() {
					Expect(db.Update(ctx, func(ctx context.Context, tx Tx) error {
						_, err = tx.CreateBucket(ctx, bucketName)
						if err != nil {
							return err
						}
						secondBucket, err = tx.Bucket(ctx, bucketName)
						if err != nil {
							return err
						}
						thirdBucket, err = tx.Bucket(ctx, bucketName)
						if err != nil {
							return err
						}
						return nil
					})).To(BeNil())
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns bucket", func() {
					Expect(bucket).NotTo(BeNil())
				})
				It("return same bucket in same tx", func() {
					Expect(secondBucket).NotTo(BeNil())
					Expect(thirdBucket).NotTo(BeNil())
					Expect(secondBucket).To(Equal(thirdBucket))
				})
			})
			Context("failed", func() {
				It("returns error", func() {
					Expect(err).NotTo(BeNil())
					Expect(errors.Is(err, BucketNotFoundError)).To(BeTrue())
				})
				It("returns no bucket", func() {
					Expect(bucket).To(BeNil())
				})
			})
		})
		Context("DeleteBucket", func() {
			JustBeforeEach(func() {
				err = db.Update(ctx, func(ctx context.Context, tx Tx) error {
					return tx.DeleteBucket(ctx, bucketName)
				})
			})
			Context("success", func() {
				BeforeEach(func() {
					Expect(db.Update(ctx, func(ctx context.Context, tx Tx) error {
						_, err = tx.CreateBucket(ctx, bucketName)
						return err
					})).To(BeNil())
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
			})
			Context("failed", func() {
				It("returns error", func() {
					Expect(err).NotTo(BeNil())
					Expect(errors.Is(err, BucketNotFoundError)).To(BeTrue())
				})
			})
		})
	})
}
