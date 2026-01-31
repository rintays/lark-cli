package larksdk

import "testing"

func TestNormalizeMailMessageID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		changed bool
	}{
		{
			name:    "empty",
			input:   "",
			want:    "",
			changed: false,
		},
		{
			name:    "already-standard-base64",
			input:   "TUlHc1NoWFhJMXgyUi9VZTNVL3h6UnlkRUdzPQ==",
			want:    "TUlHc1NoWFhJMXgyUi9VZTNVL3h6UnlkRUdzPQ==",
			changed: false,
		},
		{
			name:    "base64url-with-padding-mismatch",
			input:   "ZmVkYTQzYzUtNzY0NC00NGRhLTg5ZDctZWNmMDY4MTI1ZDA0===",
			want:    "ZmVkYTQzYzUtNzY0NC00NGRhLTg5ZDctZWNmMDY4MTI1ZDA0",
			changed: true,
		},
		{
			name:    "base64url-characters",
			input:   "ab-_",
			want:    "ab+/",
			changed: true,
		},
		{
			name:    "base64url-no-padding",
			input:   "YWJjZA",
			want:    "YWJjZA==",
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := normalizeMailMessageID(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeMailMessageID(%q)=%q, want %q", tt.input, got, tt.want)
			}
			if changed != tt.changed {
				t.Fatalf("normalizeMailMessageID(%q) changed=%t, want %t", tt.input, changed, tt.changed)
			}
		})
	}
}
