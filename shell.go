package main

import (
	"fmt"
	"strings"
)

func fakeShell(cmd string) string {
	args := strings.Fields(cmd)
	if len(args) > 0 {
		return fmt.Sprintf("-bash: %s: command not found\n", args[0])
	}

	return ""
}
