// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http"

	"github.com/gorilla/mux"
)

var globalRouter = mux.NewRouter()

func init() {
	globalRouter.StrictSlash(true)
	http.Handle("/", globalRouter)
}

// Router returns the global mux.Router used for handling all requests.
func Router() *mux.Router {
	return globalRouter
}

// AddRoute is a helper for registering a handler for an http path (route).
// Note that it adds an HSTS header to the response.
func AddRoute(route, name string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return globalRouter.HandleFunc(route, WrapHSTS(handler)).Name(name)
}

// WrapHSTS wraps the given handler func in one that sets the
// Strict-Transport-Security header on the response.
func WrapHSTS(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := "max-age=31536000; preload"
		w.Header().Add("Strict-Transport-Security", value)
		h(w, r)
	})
}
