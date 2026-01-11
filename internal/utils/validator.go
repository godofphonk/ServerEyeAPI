package utils

import (
	"errors"
	"regexp"
)

// ValidateSecretKey validates secret key format
func ValidateSecretKey(key string) error {
	if len(key) < 10 {
		return errors.New("secret key must be at least 10 characters")
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, key)
	if !matched {
		return errors.New("secret key can only contain alphanumeric characters, underscore, and dash")
	}

	return nil
}

// ValidateServerID validates server ID format
func ValidateServerID(id string) error {
	if len(id) < 5 {
		return errors.New("server ID must be at least 5 characters")
	}

	matched, _ := regexp.MatchString(`^srv_[a-zA-Z0-9]+$`, id)
	if !matched {
		return errors.New("server ID must start with 'srv_' followed by alphanumeric characters")
	}

	return nil
}

// ValidateServerKey validates server key format
func ValidateServerKey(key string) error {
	if len(key) < 10 {
		return errors.New("server key must be at least 10 characters")
	}

	matched, _ := regexp.MatchString(`^key_[a-zA-Z0-9]+$`, key)
	if !matched {
		return errors.New("server key must start with 'key_' followed by alphanumeric characters")
	}

	return nil
}
