package sm64ds

import (
	"fmt"
	"os"
)

// Die prints an error and exits (helper for the cmd/ tools).
func Die(err error) {
	fmt.Fprintln(os.Stderr, "sm64ds:", err)
	os.Exit(1)
}
