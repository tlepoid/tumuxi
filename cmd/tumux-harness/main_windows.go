//go:build windows

package main

import (
	"fmt"
	"os"
)

func main() {
	_, _ = fmt.Fprintln(os.Stderr, "tumux-harness is not supported on Windows. It requires tmux and is supported on Linux/macOS.")
	os.Exit(1)
}
