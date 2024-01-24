//go:build race
// +build race

// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package poll

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/web-platform-tests/wpt.fyi/api/query"
	"github.com/web-platform-tests/wpt.fyi/shared"
)

func TestDataRaceMetadataCache(t *testing.T) {
	versionACache := map[string][]byte{
		"original": nil,
	}
	versionBCache := map[string][]byte{
		"updated": nil,
	}

	query.MetadataMapCached = versionACache
	var readerWg sync.WaitGroup
	var writerWg sync.WaitGroup
	numberOfReaders := 200
	readersStartedCh := make(chan struct{}, numberOfReaders)
	readerWg.Add(numberOfReaders)
	writerWg.Add(1)
	// Simulate readers. Such as concurrent requests to read the cache
	for i := 0; i < numberOfReaders; i++ {
		go func() {
			defer readerWg.Done()
			time.Sleep(10 * time.Millisecond)
			cache := query.MetadataMapCached
			if !reflect.DeepEqual(cache, versionACache) && !reflect.DeepEqual(cache, versionBCache) {
				t.Log("unexpected value") // should never get here but want to read the value
			}
			readersStartedCh <- struct{}{}
		}()
	}
	// Simulate the single polling goroutine to keep the cache updated
	go func() {
		defer writerWg.Done()
		<-readersStartedCh // wait until at least one reader is finished
		// Simulate polling and updating the values
		for i := 0; i < 20; i++ {
			if i%2 == 0 {
				query.MetadataMapCached = versionBCache
			} else {
				query.MetadataMapCached = versionACache
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	readerWg.Wait()
	writerWg.Wait()
}

func TestDataRaceWebFeaturesData(t *testing.T) {
	versionACache := shared.WebFeaturesData{
		"original": {"foo": nil},
	}
	versionBCache := shared.WebFeaturesData{
		"updated": {"bar": nil},
	}
	query.SetWebFeaturesDataCache(versionACache)
	var readerWg sync.WaitGroup
	var writerWg sync.WaitGroup
	numberOfReaders := 200
	readersStartedCh := make(chan struct{}, numberOfReaders)
	readerWg.Add(numberOfReaders)
	writerWg.Add(1)
	// Simulate readers. Such as concurrent requests to read the cache
	for i := 0; i < numberOfReaders; i++ {
		go func() {
			defer readerWg.Done()
			time.Sleep(10 * time.Millisecond)
			cache := query.GetWebFeaturesDataCache()
			if !reflect.DeepEqual(cache, versionACache) && !reflect.DeepEqual(cache, versionBCache) {
				t.Log("unexpected value") // should never get here but want to read the value
			}
			readersStartedCh <- struct{}{}
		}()
	}
	// Simulate the single polling goroutine to keep the cache updated
	go func() {
		defer writerWg.Done()
		<-readersStartedCh // wait until at least one reader is finished
		// Simulate polling and updating the values
		for i := 0; i < 20; i++ {
			if i%2 == 0 {
				query.SetWebFeaturesDataCache(versionBCache)
			} else {
				query.SetWebFeaturesDataCache(versionACache)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	readerWg.Wait()
	writerWg.Wait()
}
