package phase

import (
	"github.com/nirantarasi/k3sctl/internal/shellutil"
)

func shellQuote(s string) string {
	return shellutil.Quote(s)
}
