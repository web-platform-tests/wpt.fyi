// Copyright 2022 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

// SecretManager is a simple interface for getting secrets.
type SecretManager interface {
	GetSecret(name string) ([]byte, error)
}

// GetUploader gets the Uploader by the given name.
func GetUploader(m SecretManager, uploader string) (Uploader, error) {
	value, err := m.GetSecret(uploader)
	if err != nil {
		return Uploader{}, err
	}
	return Uploader{
		Username: uploader,
		Password: string(value),
	}, nil
}
