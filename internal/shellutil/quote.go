package shellutil

import "strings"

// Quote wraps a value in single quotes with proper escaping for use in
// shell commands. Interior single quotes are escaped as '\'' which ends
// the current quoted string, inserts a literal quote via backslash, and
// reopens the quoted string.
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
