package sftp

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"/tmp/plain.sh", "'/tmp/plain.sh'"},
		{"/tmp/with space.sh", "'/tmp/with space.sh'"},
		{"/tmp/semi;rm -rf.sh", "'/tmp/semi;rm -rf.sh'"},
		{"/tmp/it's.sh", `'/tmp/it'\''s.sh'`},
	}
	for _, tt := range tests {
		if got := shellQuote(tt.in); got != tt.want {
			t.Errorf("shellQuote(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}
