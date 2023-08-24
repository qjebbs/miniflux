// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

func runScheduler(store *storage.Storage, pool *worker.Pool) {
	logger.Info(`Starting background scheduler...`)

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
	)

	if config.Opts.HasCacheService() {
		go cacheScheduler(store, config.Opts.CacheFrequency())
	}
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize int) {
	for range time.Tick(time.Duration(frequency) * time.Minute) {
		jobs, err := store.NewBatch(batchSize)
		logger.Info("[Scheduler:Feed] Pushing %d jobs to the queue", len(jobs))
		if err != nil {
			logger.Error("[Scheduler:Feed] %v", err)
		} else {
			pool.Push(jobs)
		}
	}
}

func cleanupScheduler(store *storage.Storage, frequency int) {
	for range time.Tick(time.Duration(frequency) * time.Hour) {
		runCleanupTasks(store)
	}
}

func cacheScheduler(store *storage.Storage, frequency int) {
	c := time.Tick(time.Duration(frequency) * time.Hour)
	for range c {
		if err := store.ValidateCaches(); err != nil {
			logger.Error("[Scheduler:ValidateCaches] %v", err)
		}
		if err := store.CacheMedias(); err != nil {
			logger.Error("[Scheduler:CacheMedias] %v", err)
		}
	}
}
