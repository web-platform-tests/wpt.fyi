// Copyright 2018 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package receiver

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

// TODO: This interface might also be useful to other APIs.

type gcs interface {
	NewWriter(bucketName, fileName, contentType, contentEncoding string) (io.WriteCloser, error)
}

type gcsImpl struct {
	ctx    context.Context
	client *storage.Client
}

func (g *gcsImpl) NewWriter(bucketName, fileName, contentType, contentEncoding string) (io.WriteCloser, error) {
	if g.client == nil {
		var err error
		g.client, err = storage.NewClient(g.ctx)
		if err != nil {
			return nil, err
		}
	}
	bucket := g.client.Bucket(bucketName)
	w := bucket.Object(fileName).NewWriter(g.ctx)
	if contentType != "" {
		w.ContentType = contentType
	}
	if contentEncoding != "" {
		w.ContentEncoding = contentEncoding
	}
	return w, nil
}

var _ gcs = (*gcsImpl)(nil)
