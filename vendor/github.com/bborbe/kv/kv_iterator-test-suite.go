// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func IteratorTestSuite(provider Provider) {
	GinkgoHelper()
	Context("Iterator", func() {
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

			err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
				{
					bucket, err := tx.CreateBucketIfNotExists(ctx, NewBucketName("aaaa"))
					Expect(err).To(BeNil())
					{
						err = bucket.Put(ctx, []byte("dk"), []byte("dv"))
						Expect(err).To(BeNil())
					}
				}
				{
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())
					{
						err = bucket.Put(ctx, []byte("dk"), []byte("dv"))
						Expect(err).To(BeNil())
					}
					{
						err = bucket.Put(ctx, []byte("bk"), []byte("bv"))
						Expect(err).To(BeNil())
					}
				}
				{
					bucket, err := tx.CreateBucketIfNotExists(ctx, NewBucketName("zzzz"))
					Expect(err).To(BeNil())
					{
						err = bucket.Put(ctx, []byte("dk"), []byte("dv"))
						Expect(err).To(BeNil())
					}
				}
				return nil
			})
			Expect(err).To(BeNil())
		})
		Context("Iterator", func() {
			It("Rewind", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.Iterator()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Rewind(); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("bk")))
					Expect(keys[1]).To(Equal([]byte("dk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("bv")))
					Expect(values[1]).To(Equal([]byte("dv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek before", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.Iterator()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ak")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("bk")))
					Expect(keys[1]).To(Equal([]byte("dk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("bv")))
					Expect(values[1]).To(Equal([]byte("dv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek at", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.Iterator()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("bk")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("bk")))
					Expect(keys[1]).To(Equal([]byte("dk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("bv")))
					Expect(values[1]).To(Equal([]byte("dv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek between", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.Iterator()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ck")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(1))
					Expect(keys[0]).To(Equal([]byte("dk")))
					Expect(values).To(HaveLen(1))
					Expect(values[0]).To(Equal([]byte("dv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek after", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.Iterator()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ek")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(0))
					return nil
				})
				Expect(err).To(BeNil())
			})
		})
		Context("ReverseIterator", func() {
			It("Rewind", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.IteratorReverse()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Rewind(); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("dk")))
					Expect(keys[1]).To(Equal([]byte("bk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("dv")))
					Expect(values[1]).To(Equal([]byte("bv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek before", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.IteratorReverse()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ek")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("dk")))
					Expect(keys[1]).To(Equal([]byte("bk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("dv")))
					Expect(values[1]).To(Equal([]byte("bv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek at", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.IteratorReverse()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("dk")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(2))
					Expect(keys[0]).To(Equal([]byte("dk")))
					Expect(keys[1]).To(Equal([]byte("bk")))
					Expect(values).To(HaveLen(2))
					Expect(values[0]).To(Equal([]byte("dv")))
					Expect(values[1]).To(Equal([]byte("bv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek between", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.IteratorReverse()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ck")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(1))
					Expect(keys[0]).To(Equal([]byte("bk")))
					Expect(values).To(HaveLen(1))
					Expect(values[0]).To(Equal([]byte("bv")))
					return nil
				})
				Expect(err).To(BeNil())
			})
			It("Seek after", func() {
				err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
					bucket, err := tx.CreateBucketIfNotExists(ctx, bucketName)
					Expect(err).To(BeNil())

					it := bucket.IteratorReverse()
					defer it.Close()
					var keys [][]byte
					var values [][]byte
					for it.Seek([]byte("ak")); it.Valid(); it.Next() {
						item := it.Item()
						keys = append(keys, item.Key())
						err = item.Value(func(val []byte) error {
							values = append(values, val)
							return nil
						})
						Expect(err).To(BeNil())
					}
					Expect(keys).To(HaveLen(0))
					Expect(values).To(HaveLen(0))
					return nil
				})
				Expect(err).To(BeNil())
			})
		})
	})
}
