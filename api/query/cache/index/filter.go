// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package index

// FUTURE: This file will contain code for composing filter functions from
// queries. These functions will be bound to a subset of Index data stored in
// an instance of index.

type index struct {
	tests      Tests
	runResults map[RunID]RunResults
}
