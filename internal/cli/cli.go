// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"flag"
	"fmt"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui/static"
	"miniflux.app/v2/internal/version"
)

const (
	flagInfoHelp            = "Show build information"
	flagVersionHelp         = "Show application version"
	flagMigrateHelp         = "Run SQL migrations"
	flagFlushSessionsHelp   = "Flush all sessions (disconnect users)"
	flagCreateAdminHelp     = "Create admin user"
	flagResetPasswordHelp   = "Reset user password"
	flagResetFeedErrorsHelp = "Clear all feed errors for all users"
	flagDebugModeHelp       = "Show debug logs"
	flagConfigFileHelp      = "Load configuration file"
	flagConfigDumpHelp      = "Print parsed configuration values"
	flagCacheHelp           = "Do the media cache job"
	flagCacheToDiskHelp     = "Move media caches from database to disk"
	flagCleanCacheHelp      = "Remove unused caches from disk and database"
	flagArchiveReadHelp     = "Archive read articles"
	flagHealthCheckHelp     = `Perform a health check on the given endpoint (the value "auto" try to guess the health check endpoint).`
	flagRefreshFeedsHelp    = "Refresh a batch of feeds and exit"
	flagRunCleanupTasksHelp = "Run cleanup tasks (delete old sessions and archives old entries)"
	flagFixCoverImagesHelp  = "Fix cover images display for old articles for a user"
)

// Parse parses command line arguments.
func Parse() {
	var (
		err                 error
		flagInfo            bool
		flagVersion         bool
		flagMigrate         bool
		flagFlushSessions   bool
		flagCreateAdmin     bool
		flagResetPassword   bool
		flagResetFeedErrors bool
		flagDebugMode       bool
		flagConfigFile      string
		flagConfigDump      bool
		flagCacheToDisk     bool
		flagCleanCache      bool
		flagCache           bool
		flagArchiveRead     bool
		flagHealthCheck     string
		flagRefreshFeeds    bool
		flagRunCleanupTasks bool
		flagFixCoverImages  int64
	)

	flag.BoolVar(&flagInfo, "info", false, flagInfoHelp)
	flag.BoolVar(&flagInfo, "i", false, flagInfoHelp)
	flag.BoolVar(&flagVersion, "version", false, flagVersionHelp)
	flag.BoolVar(&flagVersion, "v", false, flagVersionHelp)
	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.BoolVar(&flagFlushSessions, "flush-sessions", false, flagFlushSessionsHelp)
	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)
	flag.BoolVar(&flagResetPassword, "reset-password", false, flagResetPasswordHelp)
	flag.BoolVar(&flagResetFeedErrors, "reset-feed-errors", false, flagResetFeedErrorsHelp)
	flag.BoolVar(&flagDebugMode, "debug", false, flagDebugModeHelp)
	flag.StringVar(&flagConfigFile, "config-file", "", flagConfigFileHelp)
	flag.StringVar(&flagConfigFile, "c", "", flagConfigFileHelp)
	flag.BoolVar(&flagConfigDump, "config-dump", false, flagConfigDumpHelp)
	flag.BoolVar(&flagCache, "cache", false, flagCacheHelp)
	flag.BoolVar(&flagCacheToDisk, "cache-to-disk", false, flagCacheToDiskHelp)
	flag.BoolVar(&flagCleanCache, "cache-clean", false, flagCleanCacheHelp)
	flag.BoolVar(&flagArchiveRead, "archive-read", false, flagArchiveReadHelp)
	flag.StringVar(&flagHealthCheck, "healthcheck", "", flagHealthCheckHelp)
	flag.BoolVar(&flagRefreshFeeds, "refresh-feeds", false, flagRefreshFeedsHelp)
	flag.BoolVar(&flagRunCleanupTasks, "run-cleanup-tasks", false, flagRunCleanupTasksHelp)
	flag.Int64Var(&flagFixCoverImages, "fix-cover", 0, flagFixCoverImagesHelp)
	flag.Parse()

	cfg := config.NewParser()

	if flagConfigFile != "" {
		config.Opts, err = cfg.ParseFile(flagConfigFile)
		if err != nil {
			logger.Fatal("%v", err)
		}
	}

	config.Opts, err = cfg.ParseEnvironmentVariables()
	if err != nil {
		logger.Fatal("%v", err)
	}

	if flagConfigDump {
		fmt.Print(config.Opts)
		return
	}

	if config.Opts.LogDateTime() {
		logger.EnableDateTime()
	}

	if flagDebugMode || config.Opts.HasDebugMode() {
		logger.EnableDebug()
	}

	if flagHealthCheck != "" {
		doHealthCheck(flagHealthCheck)
		return
	}

	if flagInfo {
		info()
		return
	}

	if flagVersion {
		fmt.Println(version.Version)
		return
	}

	if config.Opts.IsDefaultDatabaseURL() {
		logger.Info("The default value for DATABASE_URL is used")
	}

	logger.Debug("Loading translations...")
	if err := locale.LoadCatalogMessages(); err != nil {
		logger.Fatal("Unable to load translations: %v", err)
	}

	logger.Debug("Loading static assets...")
	if err := static.CalculateBinaryFileChecksums(); err != nil {
		logger.Fatal("Unable to calculate binary files checksum: %v", err)
	}

	if err := static.GenerateStylesheetsBundles(); err != nil {
		logger.Fatal("Unable to generate Stylesheet bundles: %v", err)
	}

	if err := static.GenerateJavascriptBundles(); err != nil {
		logger.Fatal("Unable to generate Javascript bundles: %v", err)
	}

	db, err := database.NewConnectionPool(
		config.Opts.DatabaseURL(),
		config.Opts.DatabaseMinConns(),
		config.Opts.DatabaseMaxConns(),
		config.Opts.DatabaseConnectionLifetime(),
	)
	if err != nil {
		logger.Fatal("Unable to initialize database connection pool: %v", err)
	}
	defer db.Close()

	store := storage.NewStorage(db)

	if err := store.Ping(); err != nil {
		logger.Fatal("Unable to connect to the database: %v", err)
	}

	if flagMigrate {
		if err := database.Migrate(db); err != nil {
			logger.Fatal(`%v`, err)
		}
		return
	}

	if flagResetFeedErrors {
		store.ResetFeedErrors()
		return
	}

	if flagFlushSessions {
		flushSessions(store)
		return
	}

	if flagCreateAdmin {
		createAdmin(store)
		return
	}

	if flagResetPassword {
		resetPassword(store)
		return
	}

	if flagCacheToDisk {
		if err = store.MoveCacheToDisk(); err != nil {
			logger.Error("%v", err)
		}
		return
	}

	if flagCleanCache {
		if err = store.CleanMediaCaches(); err != nil {
			logger.Error("%v", err)
		}
		return
	}

	if flagCache {
		if err = store.ValidateCaches(); err != nil {
			logger.Error("%v", err)
		}
		if err = store.CacheMedias(); err != nil {
			logger.Error("%v", err)
		}
		return
	}

	if flagFixCoverImages > 0 {
		if err = fixCovers(store, flagFixCoverImages); err != nil {
			logger.Error("%v", err)
		}
		return
	}

	if flagArchiveRead {
		if rowsAffected, err := store.ArchiveEntries(model.EntryStatusRead, config.Opts.CleanupArchiveReadDays(), config.Opts.CleanupArchiveBatchSize()); err != nil {
			logger.Error("%v", err)
		} else {
			logger.Info("[ArchiveEntries] %d entries changed", rowsAffected)
		}
		if rowsAffected, err := store.ArchiveEntries(model.EntryStatusRead, config.Opts.CleanupArchiveUnreadDays(), config.Opts.CleanupArchiveBatchSize()); err != nil {
			logger.Error("%v", err)
		} else {
			logger.Info("[ArchiveEntries] %d entries changed", rowsAffected)
		}
		return
	}

	// Run migrations and start the daemon.
	if config.Opts.RunMigrations() {
		if err := database.Migrate(db); err != nil {
			logger.Fatal(`%v`, err)
		}
	}

	if err := database.IsSchemaUpToDate(db); err != nil {
		logger.Fatal(`You must run the SQL migrations, %v`, err)
	}

	// Create admin user and start the daemon.
	if config.Opts.CreateAdmin() {
		createAdmin(store)
	}

	if flagRefreshFeeds {
		refreshFeeds(store)
		return
	}

	if flagRunCleanupTasks {
		runCleanupTasks(store)
		return
	}

	startDaemon(store)
}
