// Package slug derives URL-safe slugs from arbitrary titles.
//
// Algorithm (must stay in lockstep with the backend):
//
//  1. Lowercase the input.
//  2. Replace each run of non-[a-z0-9] characters with a single "-".
//  3. Trim leading and trailing "-".
//
// An empty result is an error — the caller must supply --slug explicitly.
package slug

import (
	"errors"
	"strings"
	"unicode"
)

// ErrEmpty is returned when the derived slug is empty.
var ErrEmpty = errors.New("derived slug is empty; pass --slug explicitly")

// From derives a slug from title.
func From(title string) (string, error) {
	var b strings.Builder
	b.Grow(len(title))
	prevDash := true // suppress leading dash
	for _, r := range strings.ToLower(title) {
		if isAlnum(r) {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if out == "" {
		return "", ErrEmpty
	}
	return out, nil
}

func isAlnum(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	// Reject anything else, including unicode letters — backend slug rule
	// is strict ASCII [a-z0-9].
	_ = unicode.IsLetter
	return false
}
