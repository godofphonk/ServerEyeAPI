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
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateServerID generates a unique server ID
func GenerateServerID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("srv_%s", hex.EncodeToString(b))
}

// GenerateServerKey generates a unique server key
func GenerateServerKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("key_%s", hex.EncodeToString(b))
}

// GenerateSecretKey generates a unique secret key
func GenerateSecretKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("secret_%s", hex.EncodeToString(b))
}
