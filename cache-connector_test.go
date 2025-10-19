// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/teamvault-utils/v4"
)

var _ = Describe("CacheConnector", func() {
	var ctx context.Context
	var err error
	var dummyConnector teamvault.Connector
	var cacheConnector teamvault.Connector

	BeforeEach(func() {
		ctx = context.Background()
		dummyConnector = teamvault.NewDummyConnector()
		cacheConnector = teamvault.NewCacheConnector(dummyConnector)
	})

	Context("Password", func() {
		var password teamvault.Password
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			password, err = cacheConnector.Password(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct password", func() {
			Expect(
				password,
			).To(Equal(teamvault.Password("LgIWz7BC2r68P9WTtVJdfFOYrpT2tv_yw95BzhzECiU=")))
		})
		It("caches password for subsequent calls", func() {
			key := teamvault.Key("key123")
			// Second call should return cached value
			password2, err2 := cacheConnector.Password(ctx, key)
			Expect(err2).To(BeNil())
			Expect(password2).To(Equal(password))
		})
	})

	Context("User", func() {
		var user teamvault.User
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			user, err = cacheConnector.User(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct user", func() {
			Expect(user).To(Equal(teamvault.User("key123")))
		})
		It("caches user for subsequent calls", func() {
			key := teamvault.Key("key123")
			// Second call should return cached value
			user2, err2 := cacheConnector.User(ctx, key)
			Expect(err2).To(BeNil())
			Expect(user2).To(Equal(user))
		})
	})

	Context("Url", func() {
		var url teamvault.Url
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			url, err = cacheConnector.Url(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct url", func() {
			Expect(url).To(Equal(teamvault.Url("dk9kTUjDqGcvPlvF0ZOovq3sBE-0_-Y62i8mlTX_g1M=")))
		})
		It("caches url for subsequent calls", func() {
			key := teamvault.Key("key123")
			// Second call should return cached value
			url2, err2 := cacheConnector.Url(ctx, key)
			Expect(err2).To(BeNil())
			Expect(url2).To(Equal(url))
		})
	})

	Context("File", func() {
		var file teamvault.File
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			file, err = cacheConnector.File(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("caches file for subsequent calls", func() {
			key := teamvault.Key("key123")
			// Second call should return cached value
			file2, err2 := cacheConnector.File(ctx, key)
			Expect(err2).To(BeNil())
			Expect(file2).To(Equal(file))
		})
	})

	Context("Concurrent Access", func() {
		It("handles concurrent Password requests without race conditions", func() {
			key := teamvault.Key("key123")
			var wg sync.WaitGroup
			const numGoroutines = 100

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					password, err := cacheConnector.Password(ctx, key)
					Expect(err).To(BeNil())
					Expect(password).NotTo(BeEmpty())
				}()
			}
			wg.Wait()
		})

		It("handles concurrent User requests without race conditions", func() {
			key := teamvault.Key("key123")
			var wg sync.WaitGroup
			const numGoroutines = 100

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					user, err := cacheConnector.User(ctx, key)
					Expect(err).To(BeNil())
					Expect(user).NotTo(BeEmpty())
				}()
			}
			wg.Wait()
		})

		It("handles concurrent Url requests without race conditions", func() {
			key := teamvault.Key("key123")
			var wg sync.WaitGroup
			const numGoroutines = 100

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					url, err := cacheConnector.Url(ctx, key)
					Expect(err).To(BeNil())
					Expect(url).NotTo(BeEmpty())
				}()
			}
			wg.Wait()
		})

		It("handles concurrent File requests without race conditions", func() {
			key := teamvault.Key("key123")
			var wg sync.WaitGroup
			const numGoroutines = 100

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					file, err := cacheConnector.File(ctx, key)
					Expect(err).To(BeNil())
					Expect(file).NotTo(BeEmpty())
				}()
			}
			wg.Wait()
		})

		It("handles mixed concurrent requests without race conditions", func() {
			key := teamvault.Key("key123")
			var wg sync.WaitGroup
			const numGoroutines = 25

			// Concurrent Password requests
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = cacheConnector.Password(ctx, key)
				}()
			}

			// Concurrent User requests
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = cacheConnector.User(ctx, key)
				}()
			}

			// Concurrent Url requests
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = cacheConnector.Url(ctx, key)
				}()
			}

			// Concurrent File requests
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _ = cacheConnector.File(ctx, key)
				}()
			}

			wg.Wait()
		})
	})
})
