package auth

import "unicode"

// ValidatePasswordPolicy enforces baseline password strength.
// Rules:
// - at least 10 chars
// - at least one lowercase, uppercase, digit, and symbol
func ValidatePasswordPolicy(raw string) bool {
	if len(raw) < 10 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, ch := range raw {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSymbol = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSymbol
}
