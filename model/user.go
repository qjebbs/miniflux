// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"errors"
	"time"

	"miniflux.app/timezone"
)

// User represents a user in the system.
type User struct {
	ID                int64             `json:"id"`
	Username          string            `json:"username"`
	Password          string            `json:"password,omitempty"`
	IsAdmin           bool              `json:"is_admin"`
	Theme             string            `json:"theme"`
	View              string            `json:"view"`
	Language          string            `json:"language"`
	Timezone          string            `json:"timezone"`
	EntryDirection    string            `json:"entry_sorting_direction"`
	EntriesPerPage    int               `json:"entries_per_page"`
	KeyboardShortcuts bool              `json:"keyboard_shortcuts"`
	ShowReadingTime   bool              `json:"show_reading_time"`
	LastLoginAt       *time.Time        `json:"last_login_at,omitempty"`
	Extra             map[string]string `json:"extra"`
	EntrySwipe        bool              `json:"entry_swipe"`
}

// NewUser returns a new User.
func NewUser() *User {
	return &User{Extra: make(map[string]string)}
}

// ValidateUserCreation validates new user.
func (u User) ValidateUserCreation() error {
	if err := u.ValidateUserLogin(); err != nil {
		return err
	}

	return u.ValidatePassword()
}

// ValidateUserModification validates user modification payload.
func (u User) ValidateUserModification() error {
	if u.Theme != "" {
		return ValidateTheme(u.Theme)
	}

	if u.View != "" {
		return ValidateView(u.View)
	}

	if u.Password != "" {
		return u.ValidatePassword()
	}

	return nil
}

// ValidateUserLogin validates user credential requirements.
func (u User) ValidateUserLogin() error {
	if u.Username == "" {
		return errors.New("The username is mandatory")
	}

	if u.Password == "" {
		return errors.New("The password is mandatory")
	}

	return nil
}

// ValidatePassword validates user password requirements.
func (u User) ValidatePassword() error {
	if u.Password != "" && len(u.Password) < 6 {
		return errors.New("The password must have at least 6 characters")
	}

	return nil
}

// UseTimezone converts last login date to the given timezone.
func (u *User) UseTimezone(tz string) {
	if u.LastLoginAt != nil {
		*u.LastLoginAt = timezone.Convert(tz, *u.LastLoginAt)
	}
}

// Users represents a list of users.
type Users []*User

// UseTimezone converts last login timestamp of all users to the given timezone.
func (u Users) UseTimezone(tz string) {
	for _, user := range u {
		user.UseTimezone(tz)
	}
}
