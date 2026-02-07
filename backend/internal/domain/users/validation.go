// This file exists to put all of the validation functions in the package in one place.

// We declare the package

package users

// imports go here

import (
	"strings"
)

// User validation

// Validate performs basic sanity checks on required fields and simple link hygiene.
func (u User) Validate() error {
	if u.ID <= 0 {
		return ErrInvalidUserData
	}

	if strings.TrimSpace(u.Username) == "" {
		return ErrInvalidUserData
	}

	if strings.TrimSpace(u.DisplayName) == "" {
		return ErrInvalidUserData
	}

	for _, link := range []string{u.PersonalLink1, u.PersonalLink2, u.PersonalLink3, u.PersonalLink4} {
		trimmedLink := strings.TrimSpace(link)
		if trimmedLink != "" {
			if strings.ContainsAny(trimmedLink, " \t\r\n") {
				return ErrInvalidUserData
			}
			if !(strings.HasPrefix(trimmedLink, "http://") || strings.HasPrefix(trimmedLink, "https://")) {
				return ErrInvalidUserData
			}
		}
	}

	return nil
}
