// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package database // import "miniflux.app/database"

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"miniflux.app/logger"
)

const schemaVersion = 23

// Migrate executes database migrations.
func Migrate(db *sql.DB) {
	var (
		versionString          string
		currentVersion         int
		currecurrentSubVersion int
		err                    error
	)

	db.QueryRow(`select version from schema_version`).Scan(&versionString)

	vers := strings.Split(versionString, ".")
	currentVersion, err = strconv.Atoi(vers[0])
	if err != nil {
		logger.Fatal("[Migrate] %v", err)
	}
	if len(vers) > 1 {
		currecurrentSubVersion, err = strconv.Atoi(vers[1])
		if err != nil {
			logger.Fatal("[Migrate] %v", err)
		}
	}
	fmt.Println("Current schema version:", versionString)
	fmt.Println("Latest schema version:", schemaVersion)

	for version := currentVersion; version <= schemaVersion; version++ {

		rawSQL := ""
		if version > currentVersion {
			rawSQL = SqlMap["schema_version_"+strconv.Itoa(version)]
			execSchema(db, rawSQL, version, 0)
		}

		subVersion := currecurrentSubVersion + 1
		subExists := true
		for subExists {
			if rawSQL, subExists = SqlMap["schema_version_"+strconv.Itoa(version)+"_"+strconv.Itoa(subVersion)]; subExists {
				execSchema(db, rawSQL, version, subVersion)
			}
			subVersion++
		}
	}
}

func execSchema(db *sql.DB, rawSQL string, version int, subVersion int) {
	schemaVersion := ""
	if subVersion > 0 {
		schemaVersion = fmt.Sprintf("%d.%d", version, subVersion)
	} else {
		schemaVersion = strconv.Itoa(version)
	}
	fmt.Println("Migrating to version:", schemaVersion)

	tx, err := db.Begin()
	if err != nil {
		logger.Fatal("[Migrate] %v", err)
	}

	_, err = tx.Exec(rawSQL)
	if err != nil {
		tx.Rollback()
		logger.Fatal("[Migrate] %v", err)
	}

	if _, err := tx.Exec(`delete from schema_version`); err != nil {
		tx.Rollback()
		logger.Fatal("[Migrate] %v", err)
	}

	if _, err := tx.Exec(`insert into schema_version (version) values($1)`, schemaVersion); err != nil {
		tx.Rollback()
		logger.Fatal("[Migrate] %v", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Fatal("[Migrate] %v", err)
	}
}
