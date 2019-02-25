// Copyright 2019 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func apiSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var sub webpush.Subscription
	if err = json.Unmarshal(body, &sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := shared.NewAppEngineContext(r)
	aeAPI := shared.NewAppEngineAPI(ctx)
	err = shared.AddSubscription(aeAPI, sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
