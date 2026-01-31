package main

import (
	"bytes"
	"fmt"
)

/* https://www.rfc-editor.org/rfc/rfc854  */
/* https://www.rfc-editor.org/rfc/rfc1572 */

func parseEnv(b []byte) map[string]string {
	env := make(map[string]string)

	payload := []byte{IAC, SB, NEW_ENVIRON}
	start := bytes.Index(b, payload)
	if start == -1 {
		return env
	}

	fmt.Println(start)
	return env
}

func deocdeEnv(b []byte, env map[string]string) {

}
