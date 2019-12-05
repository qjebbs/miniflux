// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scheduler // import "miniflux.app/service/scheduler"

import (
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/worker"
)

// Serve starts the internal scheduler.
func Serve(store *storage.Storage, pool *worker.Pool) {
	logger.Info(`Starting scheduler...`)
	go store.CreateMediasRunOnce()
	go feedScheduler(
		store,
		pool,
		config.Opts.PollingFrequency(),
		config.Opts.BatchSize(),
	)

	go cleanupScheduler(
		store,
		config.Opts.CleanupFrequencyHours(),
		config.Opts.CleanupArchiveReadDays(),
		config.Opts.CleanupRemoveSessionsDays(),
	)
	if config.Opts.HasCacheService() {
		go cacheScheduler(store, config.Opts.CacheFrequency())
	}
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize int) {
	c := time.Tick(time.Duration(frequency) * time.Minute)
	for range c {
		jobs, err := store.NewBatch(batchSize)
		if err != nil {
			logger.Error("[Scheduler:Feed] %v", err)
		} else {
			logger.Debug("[Scheduler:Feed] Pushing %d jobs", len(jobs))
			pool.Push(jobs)
		}
	}
}

func cleanupScheduler(store *storage.Storage, frequency int, archiveDays int, sessionsDays int) {
	c := time.Tick(time.Duration(frequency) * time.Hour)
	for range c {
		nbSessions := store.CleanOldSessions(sessionsDays)
		nbUserSessions := store.CleanOldUserSessions(sessionsDays)
		logger.Info("[Scheduler:Cleanup] Cleaned %d sessions and %d user sessions", nbSessions, nbUserSessions)

		if err := store.ArchiveEntries(archiveDays); err != nil {
			logger.Error("[Scheduler:Cleanup:ArchiveEntries] %v", err)
		}
		// Important: clean caches before media, or caches in disk will be orphan files.
		if err := store.CleanMediaCaches(); err != nil {
			logger.Error("[Scheduler:Cleanup:Caches] %v", err)
		}
		if err := store.CleanupMedias(); err != nil {
			logger.Error("[Scheduler:Cleanup:Medias] %v", err)
		}
	}
}

func cacheScheduler(store *storage.Storage, frequency int) {
	c := time.Tick(time.Duration(frequency) * time.Hour)
	for range c {
		if err := store.CacheMedias(30); err != nil {
			logger.Error("[Scheduler:Cache] %v", err)
		}
	}
}
