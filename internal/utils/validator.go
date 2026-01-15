// Copyright (c) 2026 godofphonk
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	if len(key) < 5 {
		return errors.New("server key must be at least 5 characters")
	}

	matched, _ := regexp.MatchString(`^key_[a-zA-Z0-9]+$`, key)
	if !matched {
		return errors.New("server key must start with 'key_' followed by alphanumeric characters")
	}

	return nil
}
