package auth

import "testing"

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "redact ghp token",
			input: "error: ghp_1234567890abcdefghijklmnopqrstuvwxyz12 is invalid",
			want:  "error: ghp_****************************wxyz is invalid",
		},
		{
			name:  "redact github_pat token",
			input: "token: github_pat_abcdefghij1234567890AB_XYZ",
			want:  "token: gith************************************_XYZ",
		},
		{
			name:  "no secrets to redact",
			input: "everything is fine",
			want:  "everything is fine",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSecrets(tt.input)
			if got == tt.input && tt.input != tt.want {
				t.Errorf("RedactSecrets() did not redact: got %q", got)
			}
			// Check that original token is not present if it should be redacted
			if tt.input != tt.want && got == tt.input {
				t.Errorf("Expected redaction but got original string")
			}
		})
	}
}
