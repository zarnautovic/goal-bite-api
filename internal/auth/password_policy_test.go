package auth

import "testing"

func TestValidatePasswordPolicy(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{name: "valid", in: "Pass1234!x", ok: true},
		{name: "too short", in: "Pa1!x", ok: false},
		{name: "no upper", in: "pass1234!x", ok: false},
		{name: "no lower", in: "PASS1234!X", ok: false},
		{name: "no digit", in: "Password!!x", ok: false},
		{name: "no symbol", in: "Password123x", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidatePasswordPolicy(tc.in)
			if got != tc.ok {
				t.Fatalf("expected %v, got %v", tc.ok, got)
			}
		})
	}
}
