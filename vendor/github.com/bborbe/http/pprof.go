// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// RegisterPprof register pprof http endpoint to gorilla mux
// https://www.codereliant.io/memory-leaks-with-pprof/
// kubectl -n erpnext port-forward service/hubspot-resource-exporter 9090:9090
// go tool pprof -alloc_space http://localhost:9090/debug/pprof/heap
// > web
// or
// go tool pprof -http=127.0.0.1:16666 -alloc_space http://localhost:9090/debug/pprof/heap
func RegisterPprof(router *mux.Router) {
	router.PathPrefix("/debug/pprof/cmdline").HandlerFunc(pprof.Cmdline)
	router.PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	router.PathPrefix("/debug/pprof/symbol").HandlerFunc(pprof.Symbol)
	router.PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
}
