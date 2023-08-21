// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/storage"

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"miniflux.app/v2/model"
)

// AnotherCategoryExists checks if another category exists with the same title.
func (s *Storage) AnotherCategoryExists(userID, categoryID int64, title string) bool {
	var result bool
	query := `SELECT true FROM categories WHERE user_id=$1 AND id != $2 AND lower(title)=lower($3) LIMIT 1`
	s.db.QueryRow(query, userID, categoryID, title).Scan(&result)
	return result
}

// CategoryTitleExists checks if the given category exists into the database.
func (s *Storage) CategoryTitleExists(userID int64, title string) bool {
	var result bool
	query := `SELECT true FROM categories WHERE user_id=$1 AND lower(title)=lower($2) LIMIT 1`
	s.db.QueryRow(query, userID, title).Scan(&result)
	return result
}

// CategoryIDExists checks if the given category exists into the database.
func (s *Storage) CategoryIDExists(userID, categoryID int64) bool {
	var result bool
	query := `SELECT true FROM categories WHERE user_id=$1 AND id=$2`
	s.db.QueryRow(query, userID, categoryID).Scan(&result)
	return result
}

// Category returns a category from the database.
func (s *Storage) Category(userID, categoryID int64) (*model.Category, error) {
	var category model.Category

	query := `SELECT id, user_id, title, view FROM categories WHERE user_id=$1 AND id=$2`
	err := s.db.QueryRow(query, userID, categoryID).Scan(&category.ID, &category.UserID, &category.Title, &category.View)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch category: %v`, err)
	default:
		return &category, nil
	}
}

// FirstCategory returns the first category for the given user.
func (s *Storage) FirstCategory(userID int64) (*model.Category, error) {
	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 ORDER BY title ASC LIMIT 1`

	var category model.Category
	err := s.db.QueryRow(query, userID).Scan(&category.ID, &category.UserID, &category.Title)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch category: %v`, err)
	default:
		return &category, nil
	}
}

// CategoryByTitle finds a category by the title.
func (s *Storage) CategoryByTitle(userID int64, title string) (*model.Category, error) {
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 AND title=$2`
	err := s.db.QueryRow(query, userID, title).Scan(&category.ID, &category.UserID, &category.Title)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch category: %v`, err)
	default:
		return &category, nil
	}
}

// Categories returns all categories that belongs to the given user.
func (s *Storage) Categories(userID int64) (model.Categories, error) {
	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 ORDER BY title ASC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch categories: %v`, err)
	}
	defer rows.Close()

	categories := make(model.Categories, 0)
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Title); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch category row: %v`, err)
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

// CategoriesWithFeedCount returns all categories with the number of feeds.
func (s *Storage) CategoriesWithFeedCount(userID int64, nsfw bool) (model.Categories, error) {
	user, err := s.UserByID(userID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			c.id,
			c.user_id,
			c.title,
			(SELECT count(*) FROM feeds WHERE feeds.category_id=c.id %s) AS count,
			(SELECT count(*)
			   FROM feeds
			     JOIN entries ON (feeds.id = entries.feed_id)
			   WHERE feeds.category_id = c.id AND entries.status = 'unread' %s) AS count_unread
		FROM categories c
		WHERE
			user_id=$1
	`

	if user.CategoriesSortingOrder == "alphabetical" {
		query = query + `
			ORDER BY
				c.title ASC
		`
	} else {
		query = query + `
			ORDER BY
				count_unread DESC,
				c.title ASC
		`
	}
	nsfwCond := ""
	if nsfw {
		nsfwCond = "AND feeds.nsfw = 'f'"
	}
	query = fmt.Sprintf(query, nsfwCond, nsfwCond)
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch categories: %v`, err)
	}
	defer rows.Close()

	categories := make(model.Categories, 0)
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Title, &category.FeedCount, &category.TotalUnread); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch category row: %v`, err)
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

// CreateCategory creates a new category.
func (s *Storage) CreateCategory(userID int64, request *model.CategoryRequest) (*model.Category, error) {
	var category model.Category

	query := `
		INSERT INTO categories
			(user_id, title, view)
		VALUES
			($1, $2, $3)
		RETURNING
			id,
			user_id,
			title
	`
	err := s.db.QueryRow(
		query,
		userID,
		request.Title,
		request.View,
	).Scan(
		&category.ID,
		&category.UserID,
		&category.Title,
	)

	if err != nil {
		return nil, fmt.Errorf(`store: unable to create category %q: %v`, request.Title, err)
	}

	return &category, nil
}

// UpdateCategory updates an existing category.
func (s *Storage) UpdateCategory(category *model.Category) error {
	query := `UPDATE categories SET title=$1, view=$2 WHERE id=$3 AND user_id=$4`
	_, err := s.db.Exec(
		query,
		category.Title,
		category.View,
		category.ID,
		category.UserID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update category: %v`, err)
	}

	return nil
}

// RemoveCategory deletes a category.
func (s *Storage) RemoveCategory(userID, categoryID int64) error {
	query := `DELETE FROM categories WHERE id in (
		SELECT c.id FROM categories c
		LEFT JOIN feeds f ON f.category_id = c.id
		WHERE c.id = $1 AND c.user_id = $2 AND f.id IS NULL
	)`
	result, err := s.db.Exec(query, categoryID, userID)
	if err != nil {
		return fmt.Errorf(`store: unable to remove this category: %v`, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store: unable to remove this category: %v`, err)
	}

	if count == 0 {
		return errors.New(`store: no category has been removed`)
	}

	return nil
}

// UpdateCategoryView updates a category view setting.
func (s *Storage) UpdateCategoryView(userID int64, categoryID int64, view string) error {
	if _, ok := model.Views()[view]; !ok {
		return fmt.Errorf("invalid view value: %v", view)
	}
	query := `UPDATE categories SET view = $3 WHERE user_id=$1 AND id=$2`
	result, err := s.db.Exec(query, userID, categoryID, view)
	if err != nil {
		return fmt.Errorf("unable to set view for category #%d: %v", categoryID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to set view for category #%d: %v", categoryID, err)
	}

	if count == 0 {
		return errors.New("nothing has been updated")
	}
	return nil
}

// RemoveAndReplaceCategoriesByName deletes the given categories, replacing those categories with the user's first
// category on affected feeds
func (s *Storage) RemoveAndReplaceCategoriesByName(userid int64, titles []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.New("unable to begin transaction")
	}

	titleParam := pq.Array(titles)
	var count int
	query := "SELECT count(*) FROM categories WHERE user_id = $1 and title != ANY($2)"
	err = tx.QueryRow(query, userid, titleParam).Scan(&count)
	if err != nil {
		tx.Rollback()
		return errors.New("unable to retrieve category count")
	}
	if count < 1 {
		tx.Rollback()
		return errors.New("at least 1 category must remain after deletion")
	}

	query = `
		WITH d_cats AS (SELECT id FROM categories WHERE user_id = $1 AND title = ANY($2)) 
		UPDATE feeds 
		 SET category_id = 
		  (SELECT id 
			FROM categories 
			WHERE user_id = $1 AND id NOT IN (SELECT id FROM d_cats) 
			ORDER BY title ASC 
			LIMIT 1) 
		WHERE user_id = $1 AND category_id IN (SELECT id FROM d_cats)
	`
	_, err = tx.Exec(query, userid, titleParam)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to replace categories: %v", err)
	}

	query = "DELETE FROM categories WHERE user_id = $1 AND title = ANY($2)"
	_, err = tx.Exec(query, userid, titleParam)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to delete categories: %v", err)
	}
	tx.Commit()
	return nil
}
