package main

// https://www.rfc-editor.org/rfc/rfc1572
const (
	NEW_ENVIRON = 39
	IS          = 0
	SEND        = 1
	VAR         = 0
	VALUE       = 1
	ESC         = 2
	USERVAR     = 3
)

func parseEnv(b []byte, env map[string]string) {
	const (
		stNone = iota
		stName
		stValue
		stEsc
	)

	if len(b) <= 2 || b[0] != NEW_ENVIRON || b[1] != IS {
		return
	}

	b = b[2:]
	var nameBuf, valBuf []byte
	state := stNone
	i := 0

	for i < len(b) {
		switch b[i] {
		case VAR, USERVAR:
			if len(nameBuf) > 0 {
				env[string(nameBuf)] = string(valBuf)
			}
			nameBuf = nil
			valBuf = nil
			state = stName
			i++

		case VALUE:
			state = stValue
			i++

		case ESC:
			i++
			if i < len(b) {
				if state == stValue {
					valBuf = append(valBuf, b[i])
				} else if state == stName {
					nameBuf = append(nameBuf, b[i])
				}
				i++
			}

		default:
			if state == stValue {
				valBuf = append(valBuf, b[i])
			} else if state == stName {
				nameBuf = append(nameBuf, b[i])
			}
			i++
		}
	}

	if len(nameBuf) > 0 {
		env[string(nameBuf)] = string(valBuf)
	}
}
