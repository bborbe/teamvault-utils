// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"runtime"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RelationStoreTestSuite(provider Provider) {
	GinkgoHelper()

	Context("RelationStore", func() {
		var ctx context.Context
		var db DB
		var err error
		var relationStore RelationStore[string, string]
		BeforeEach(func() {
			ctx = context.Background()
			db, err = provider.Get(ctx)
			Expect(db).NotTo(BeNil())
			Expect(err).To(BeNil())

			relationStore = NewRelationStore[string, string](db, "test")
		})
		AfterEach(func() {
			_ = db.Close()
		})
		Context("IDs", func() {
			var ids []string
			Context("empty", func() {
				BeforeEach(func() {
					ids, err = relationStore.IDs(ctx, "abc")
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns no ids", func() {
					Expect(ids).To(HaveLen(0))
				})
			})
		})
		Context("Invert", func() {
			BeforeEach(func() {
				Expect(relationStore.Add(ctx, "k1", []string{"v1", "v2"})).To(BeNil())
			})
			It("returns ids", func() {
				values, err := relationStore.RelatedIDs(ctx, "k1")
				Expect(err).To(BeNil())
				Expect(values).To(Equal([]string{"v1", "v2"}))
			})
			It("returns inverted ids", func() {
				values, err := relationStore.Invert().RelatedIDs(ctx, "v1")
				Expect(err).To(BeNil())
				Expect(values).To(Equal([]string{"k1"}))
			})
		})
		Context("RelatedIDs", func() {
			var ids []string
			Context("empty", func() {
				BeforeEach(func() {
					ids, err = relationStore.RelatedIDs(ctx, "abc")
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns no ids", func() {
					Expect(ids).To(HaveLen(0))
				})
			})
		})
		Context("Add", func() {
			BeforeEach(func() {
				err = relationStore.Add(ctx, "c1", []string{"i1", "i2"})
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("returns IDs", func() {
				ids, err := relationStore.IDs(ctx, "i1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(1))
				Expect(ids[0]).To(Equal("c1"))
			})
			It("returns RelatedIDs", func() {
				ids, err := relationStore.RelatedIDs(ctx, "c1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(2))
				Expect(ids[0]).To(Equal("i1"))
				Expect(ids[1]).To(Equal("i2"))
			})
		})
		Context("Replace", func() {
			var data map[string][]string
			BeforeEach(func() {
				data = make(map[string][]string)
			})

			BeforeEach(func() {
				for k, v := range data {
					Expect(relationStore.Add(ctx, k, v)).To(BeNil())
				}
				err = relationStore.Replace(ctx, "c1", []string{"i1", "i2"})
			})
			Context("without data", func() {
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns IDs", func() {
					ids, err := relationStore.IDs(ctx, "i1")
					Expect(err).To(BeNil())
					Expect(ids).To(HaveLen(1))
					Expect(ids[0]).To(Equal("c1"))
				})
				It("returns RelatedIDs", func() {
					ids, err := relationStore.RelatedIDs(ctx, "c1")
					Expect(err).To(BeNil())
					Expect(ids).To(HaveLen(2))
					Expect(ids[0]).To(Equal("i1"))
					Expect(ids[1]).To(Equal("i2"))
				})
			})
			Context("with data", func() {
				BeforeEach(func() {
					data["c1"] = []string{"banana"}
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("returns IDs", func() {
					ids, err := relationStore.IDs(ctx, "i1")
					Expect(err).To(BeNil())
					Expect(ids).To(HaveLen(1))
					Expect(ids[0]).To(Equal("c1"))
				})
				It("returns RelatedIDs", func() {
					ids, err := relationStore.RelatedIDs(ctx, "c1")
					Expect(err).To(BeNil())
					Expect(ids).To(HaveLen(2))
					Expect(ids[0]).To(Equal("i1"))
					Expect(ids[1]).To(Equal("i2"))
				})
			})
		})
		Context("Remove", func() {
			BeforeEach(func() {
				err = relationStore.Add(ctx, "c1", []string{"i1", "i2", "i3"})
				Expect(err).To(BeNil())
				err = relationStore.Remove(ctx, "c1", []string{"i3"})
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("returns IDs", func() {
				ids, err := relationStore.IDs(ctx, "i1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(1))
				Expect(ids[0]).To(Equal("c1"))
			})
			It("returns RelatedIDs", func() {
				ids, err := relationStore.RelatedIDs(ctx, "c1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(2))
				Expect(ids[0]).To(Equal("i1"))
				Expect(ids[1]).To(Equal("i2"))
			})
		})
		Context("Delete", func() {
			BeforeEach(func() {
				Expect(relationStore.Add(ctx, "c1", []string{"i1", "i2"})).To(BeNil())
				Expect(relationStore.Add(ctx, "c2", []string{"i1", "i2"})).To(BeNil())
				err = relationStore.Delete(ctx, "c1")
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("returns IDs", func() {
				ids, err := relationStore.IDs(ctx, "i1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(1))
				Expect(ids[0]).To(Equal("c2"))
			})
			It("returns RelatedIDs of c1", func() {
				ids, err := relationStore.RelatedIDs(ctx, "c1")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(0))
			})
			It("returns RelatedIDs of c2", func() {
				ids, err := relationStore.RelatedIDs(ctx, "c2")
				Expect(err).To(BeNil())
				Expect(ids).To(HaveLen(2))
			})
		})
		Context("StreamIDs", func() {
			var result []string
			JustBeforeEach(func() {
				result = []string{}
				ctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				ch := make(chan string, runtime.NumCPU())
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							return
						case r, ok := <-ch:
							if !ok {
								return
							}
							result = append(result, r)
						}
					}
				}()
				err = relationStore.StreamIDs(ctx, ch)
				close(ch)
				wg.Wait()
			})
			Context("empty", func() {
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
			})
			Context("one", func() {
				BeforeEach(func() {
					err = relationStore.Add(ctx, "a", []string{"1", "2", "3"})
					Expect(err).To(BeNil())
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("return one id", func() {
					Expect(result).To(HaveLen(1))
					Expect(result[0]).To(Equal("a"))
				})
			})
		})
		Context("StreamRelatedIDs", func() {
			var result []string
			JustBeforeEach(func() {
				result = []string{}
				ctx, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				ch := make(chan string, runtime.NumCPU())
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							return
						case r, ok := <-ch:
							if !ok {
								return
							}
							result = append(result, r)
						}
					}
				}()
				err = relationStore.StreamRelatedIDs(ctx, ch)
				close(ch)
				wg.Wait()
			})
			Context("empty", func() {
				BeforeEach(func() {
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
			})
			Context("one", func() {
				BeforeEach(func() {
					err = relationStore.Add(ctx, "a", []string{"1", "2", "3"})
					Expect(err).To(BeNil())
				})
				It("returns no error", func() {
					Expect(err).To(BeNil())
				})
				It("return three id", func() {
					Expect(result).To(HaveLen(3))
					Expect(result[0]).To(Equal("1"))
					Expect(result[1]).To(Equal("2"))
					Expect(result[2]).To(Equal("3"))
				})
			})
		})
	})
}
