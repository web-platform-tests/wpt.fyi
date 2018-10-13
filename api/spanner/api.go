// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package spanner

type API struct {
	Authenticator

	ProjectID          string
	Database           string
	GCPCredentialsFile *string
}
