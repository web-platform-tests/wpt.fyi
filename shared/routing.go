// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var globalRouter *mux.Router

// Router returns the global mux.Router used for handling all requests.
func Router() *mux.Router {
	if globalRouter == nil {
		globalRouter = mux.NewRouter()
		globalRouter.StrictSlash(true)
		http.Handle("/", globalRouter)
	}
	return globalRouter
}

// AddRoute is a helper for registering a handler for an http path (route).
// Note that it adds an HSTS header to the response.
func AddRoute(route, name string, h http.HandlerFunc) *mux.Route {
	return Router().Handle(route, WrapHSTS(h)).Name(name)
}

// WrapHSTS wraps the given handler func in one that sets the
// Strict-Transport-Security header on the response.
func WrapHSTS(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := "max-age=31536000; preload"
		w.Header().Add("Strict-Transport-Security", value)
		h.ServeHTTP(w, r)
	})
}

// WrapPermissiveCORS wraps the given handler func in one that sets an
// all-permissive CORS header on the response.
func WrapPermissiveCORS(h http.HandlerFunc) http.HandlerFunc {
	cors := handlers.CORS().
		AllowedOrigins([]string{"*"})
	return cors(h).ServeHTTP
}

// WrapApplicationJSON wraps the given handler func in one that sets a Content-Type
// header of "text/json" on the response.
func WrapApplicationJSON(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}
