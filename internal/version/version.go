package version

import (
	"fmt"
	"strings"
)

// Version is set at build time; default is dev.
var Version = "dev"

// Commit is set at build time; default is empty.
var Commit = ""

// Date is set at build time; default is empty.
var Date = ""

// String returns a human-friendly version string.
func String() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}
	c := strings.TrimSpace(Commit)
	d := strings.TrimSpace(Date)
	if c == "" && d == "" {
		return v
	}
	if c == "" {
		return fmt.Sprintf("%s (%s)", v, d)
	}
	if d == "" {
		return fmt.Sprintf("%s (%s)", v, c)
	}
	return fmt.Sprintf("%s (%s %s)", v, c, d)
}
