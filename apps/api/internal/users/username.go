// Package users owns user accounts: the username rules and first-username
// generation in this file, plus (in sibling files) the store-backed settings
// API. This file is deliberately dependency-free — no database, no HTTP —
// mirroring the framework-free internal packages (mailparse, htmltext) so the
// rules can be exhaustively unit-tested in isolation.
package users

import (
	"errors"
	"regexp"
	"strings"
)

// Username length bounds. A username is also the user's display name; there is
// no separate display-name field.
const (
	MinLen = 3
	MaxLen = 30
)

// Validate rejects a username. Callers map these to a 400:
//   - ErrUsernameInvalid: wrong length, illegal character, or a non-alphanumeric
//     first/last character.
//   - ErrUsernameReserved: a name the site keeps for routes or to block
//     impersonation (see Reserved).
var (
	ErrUsernameInvalid  = errors.New("username: invalid format")
	ErrUsernameReserved = errors.New("username: reserved")
)

// usernameRE encodes the whole format rule in one pass: 3–30 characters drawn
// from ASCII letters, digits, hyphen and underscore, starting AND ending with an
// alphanumeric. The {1,28} middle plus the two anchored ends yields the 3–30
// length bound. The charset is ASCII-only on purpose: the uniqueness index is
// COLLATE NOCASE, which folds ASCII only, so restricting input here keeps
// "JaneDoe" and "janedoe" provably the same handle.
var usernameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{1,28}[A-Za-z0-9]$`)

// Validate reports whether s is an acceptable username, returning a specific
// sentinel error otherwise (see the error vars). It never partially accepts:
// Generate is guaranteed to produce a value that passes this check.
func Validate(s string) error {
	if !usernameRE.MatchString(s) {
		return ErrUsernameInvalid
	}
	if IsReserved(s) {
		return ErrUsernameReserved
	}
	return nil
}

// reserved is the case-insensitive set of names nobody may take. Two categories:
// paths the site already serves (or plausibly will), so a future /u/<username>/
// route can never shadow or be shadowed by a real page; and identities that
// would let one user impersonate staff or the brand. Stored lowercase; look up
// via IsReserved, which folds case. Adding a name here is the cheap insurance
// that makes public profile URLs safe later without a data migration.
var reserved = map[string]struct{}{
	// Route collisions (present or plausible).
	"about": {}, "events": {}, "settings": {}, "auth": {}, "api": {},
	"login": {}, "logout": {}, "me": {}, "s": {}, "learn": {}, "jobs": {},
	"forum": {}, "puzzles": {}, "tools": {}, "codewars": {}, "contact": {},
	"credits": {}, "discounts": {}, "404": {}, "u": {},
	// Impersonation.
	"admin": {}, "root": {}, "staff": {}, "moderator": {}, "support": {},
	"system": {}, "codeselfstudy": {},
}

// IsReserved reports whether s is a reserved name, case-insensitively.
func IsReserved(s string) bool {
	_, ok := reserved[strings.ToLower(s)]
	return ok
}

// Generate derives a first username for a brand-new account from the WorkOS
// profile. It always returns a value that satisfies Validate — callers can seed
// the row with it directly. It does NOT guarantee uniqueness; the store owns the
// -2/-3 dedupe against the unique index, because uniqueness is a database
// property, not a string property.
//
// Preference order: the slugified full name, then the email local-part, then the
// literal "user". A reserved result (e.g. "admin") is disambiguated with a
// trailing digit rather than discarded. Note the email fallback can surface part
// of the address in a name that may become public later; the sign-up UX
// mitigates this by showing the generated name in a focused, pre-selected box
// before it is ever exposed.
func Generate(firstName, lastName, email string) string {
	for _, raw := range []string{firstName + lastName, localPart(email)} {
		s := slug(raw)
		if len(s) > MaxLen {
			s = s[:MaxLen]
		}
		if len(s) < MinLen {
			continue
		}
		if IsReserved(s) {
			s = s + "1" // "admin" -> "admin1"; slug output is short, stays under MaxLen
		}
		if Validate(s) == nil {
			return s
		}
	}
	return "user"
}

// slug lowercases s and keeps only ASCII alphanumerics, dropping everything else
// (spaces, dots, punctuation). "Jane Doe" -> "janedoe", "jane.doe" -> "janedoe".
// The result is therefore always a valid username body (its every character is
// alphanumeric, so any non-empty, length-bounded slice passes usernameRE).
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// localPart returns the portion of email before the first '@', or the whole
// string when there is no '@'.
func localPart(email string) string {
	if i := strings.IndexByte(email, '@'); i >= 0 {
		return email[:i]
	}
	return email
}
