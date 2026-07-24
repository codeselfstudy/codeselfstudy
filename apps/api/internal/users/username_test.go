package users

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want error
	}{
		// Accepted.
		{"min length", "abc", nil},
		{"with underscore", "jane_doe", nil},
		{"with hyphen", "a-b", nil},
		{"digits", "user123", nil},
		{"mixed case preserved", "JaneDoe", nil},
		{"max length", strings.Repeat("a", MaxLen), nil},
		{"digits at ends around symbols", "1_-_2", nil},

		// Format failures.
		{"empty", "", ErrUsernameInvalid},
		{"too short", "ab", ErrUsernameInvalid},
		{"too long", strings.Repeat("a", MaxLen+1), ErrUsernameInvalid},
		{"illegal char", "ab!c", ErrUsernameInvalid},
		{"space", "ab c", ErrUsernameInvalid},
		{"leading hyphen", "-abc", ErrUsernameInvalid},
		{"trailing hyphen", "abc-", ErrUsernameInvalid},
		{"leading underscore", "_abc", ErrUsernameInvalid},
		{"trailing underscore", "abc_", ErrUsernameInvalid},
		{"unicode", "jösé", ErrUsernameInvalid},

		// Reserved (format is fine).
		{"reserved admin", "admin", ErrUsernameReserved},
		{"reserved admin uppercase", "ADMIN", ErrUsernameReserved},
		{"reserved settings mixed", "Settings", ErrUsernameReserved},
		{"reserved api", "api", ErrUsernameReserved},
		{"reserved brand", "codeselfstudy", ErrUsernameReserved},
		{"reserved profile prefix", "u", ErrUsernameInvalid}, // also too short, format wins
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Validate(tc.in); !errors.Is(got, tc.want) {
				t.Errorf("Validate(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// Every reserved name (long enough to clear the format check) must be rejected
// as reserved, in any case, so the set can't silently rot as routes are added.
func TestReservedNamesRejected(t *testing.T) {
	for name := range reserved {
		if len(name) < MinLen {
			continue // e.g. "s", "u", "me" fail the format check first; covered above
		}
		for _, variant := range []string{name, strings.ToUpper(name), capitalize(name)} {
			if err := Validate(variant); !errors.Is(err, ErrUsernameReserved) {
				t.Errorf("Validate(%q) = %v, want ErrUsernameReserved", variant, err)
			}
		}
	}
}

// capitalize upper-cases the first byte only — a mixed-case variant for the
// case-insensitivity check, without the deprecated strings.Title.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func TestGenerate(t *testing.T) {
	cases := []struct {
		name                       string
		first, last, email, expect string
	}{
		{"full name", "Jane", "Doe", "jane.doe@example.com", "janedoe"},
		{"name with spaces and punctuation", "Ada", "O'Lovelace", "a@b.com", "adaolovelace"},
		{"uppercased name is folded", "JANE", "DOE", "x@y.com", "janedoe"},
		{"falls back to email local part", "", "", "coder99@example.com", "coder99"},
		{"email local part is slugified", "", "", "jane.doe@example.com", "janedoe"},
		{"short name falls through to email", "Al", "", "bighandle@example.com", "bighandle"},
		{"everything empty yields user", "", "", "", "user"},
		{"unusable email yields user", "", "", "@nope.com", "user"},
		{"reserved slug is disambiguated", "Admin", "", "x@y.com", "admin1"},
		{"name too short and email too short yields user", "Jo", "", "al@x.com", "user"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate(tc.first, tc.last, tc.email)
			if got != tc.expect {
				t.Errorf("Generate(%q,%q,%q) = %q, want %q", tc.first, tc.last, tc.email, got, tc.expect)
			}
		})
	}
}

// Whatever Generate returns must itself be a valid username — the sign-up flow
// seeds the row with it directly, so an invalid generation would be a stored
// broken account.
func TestGenerateAlwaysValid(t *testing.T) {
	cases := [][3]string{
		{"Jane", "Doe", "jane@example.com"},
		{"", "", ""},
		{"", "", "@bad.com"},
		{"Admin", "", "root@example.com"},
		{"A", "b", "c@d.com"},
		{strings.Repeat("Verylongname", 5), "", "x@y.com"},
		{"你好", "世界", "ünïcodé@example.com"},
		{"System", "", "support@example.com"},
	}
	for _, c := range cases {
		got := Generate(c[0], c[1], c[2])
		if err := Validate(got); err != nil {
			t.Errorf("Generate(%q,%q,%q) = %q, which fails Validate: %v", c[0], c[1], c[2], got, err)
		}
		if len(got) > MaxLen {
			t.Errorf("Generate(%q,%q,%q) = %q exceeds MaxLen %d", c[0], c[1], c[2], got, MaxLen)
		}
	}
}
