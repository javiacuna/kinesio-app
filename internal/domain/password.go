package domain

import "unicode"

// PasswordPolicyViolation returns the first policy rule a password breaks
// ("min_8_chars", "missing_uppercase", "missing_lowercase", "missing_number",
// "missing_special"), or "" if the password satisfies the policy.
func PasswordPolicyViolation(password string) string {
	if len(password) < 8 {
		return "min_8_chars"
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}

	switch {
	case !hasUpper:
		return "missing_uppercase"
	case !hasLower:
		return "missing_lowercase"
	case !hasDigit:
		return "missing_number"
	case !hasSpecial:
		return "missing_special"
	default:
		return ""
	}
}
