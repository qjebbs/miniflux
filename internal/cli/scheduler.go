// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"log/slog"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"
)

func runScheduler(store *storage.Storage, pool *worker.Pool) {
	slog.Debug(`Starting background scheduler...`)

	go store.CreateMediasRunOnce()

	go feedScheduler(
		store,
		pool,
		config.Opts.PollingFrequency(),
		config.Opts.BatchSize(),
		config.Opts.PollingParsingErrorLimit(),
	)

	go cleanupScheduler(
		store,
		config.Opts.CleanupFrequencyHours(),
	)

	if config.Opts.HasCacheService() {
		go cacheScheduler(store, config.Opts.CacheFrequency())
	}
}

func feedScheduler(store *storage.Storage, pool *worker.Pool, frequency, batchSize, errorLimit int) {
	for range time.Tick(time.Duration(frequency) * time.Minute) {
		// Generate a batch of feeds for any user that has feeds to refresh.
		batchBuilder := store.NewBatchBuilder()
		batchBuilder.WithBatchSize(batchSize)
		batchBuilder.WithErrorLimit(errorLimit)
		batchBuilder.WithoutDisabledFeeds()
		batchBuilder.WithNextCheckExpired()

		if jobs, err := batchBuilder.FetchJobs(); err != nil {
			slog.Error("Unable to fetch jobs from database", slog.Any("error", err))
		} else if len(jobs) > 0 {
			slog.Info("Created a batch of feeds",
				slog.Int("nb_jobs", len(jobs)),
			)
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
			slog.Error("scheduler: unable to validate csaches]", slog.Any("error", err))
		}
		if err := store.CacheMedias(); err != nil {
			slog.Error("scheduler: unable to cache medias", slog.Any("error", err))
		}
	}
}
