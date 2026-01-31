package main

import (
	"bytes"
	"io"
	"net"
	"time"
)

// https://www.rfc-editor.org/rfc/rfc854
// https://www.rfc-editor.org/rfc/rfc1572
const (
	IAC  = 255
	WILL = 251
	WONT = 252
	DO   = 253
	DONT = 254
	SB   = 250
	SE   = 240
	ECHO = 1
)

type TelnetReader struct {
	r   io.Reader
	Env map[string]string

	state   int
	sb      bytes.Buffer
	pending *byte
}

const (
	stNormal = iota
	stIAC
	stIAC_IGNORE
	stSB
	stSB_IAC
)

func NewTelnetReader(r io.Reader) *TelnetReader {
	return &TelnetReader{
		Env: make(map[string]string),
		r:   r,
	}
}

func (t *TelnetReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if t.pending != nil {
		p[0] = *t.pending
		t.pending = nil
		return 1, nil
	}

	buf := []byte{0}
	// Read byte by byte
	for {
		_, err := t.r.Read(buf)
		if err != nil {
			return 0, err
		}
		b := buf[0]

		switch t.state {
		case stNormal:
			// IAC telnet command
			if b == IAC {
				t.state = stIAC
				continue
			}
			p[0] = b
			return 1, nil

		case stIAC:
			switch b {
			// IAC escaped
			case IAC:
				t.state = stNormal
				p[0] = IAC
				return 1, nil

			// start recording in sb buffer
			case SB:
				t.sb.Reset()
				t.state = stSB

			case WILL, WONT, DO, DONT:
				t.state = stIAC_IGNORE

			// ignore other then sb
			default:
				t.state = stNormal
			}

		case stIAC_IGNORE:
			t.state = stNormal

		case stSB:
			if b == IAC {
				t.state = stSB_IAC
				continue
			}
			t.sb.WriteByte(b)

		case stSB_IAC:
			switch b {
			case SE:
				parseEnv(t.sb.Bytes(), t.Env)
				t.state = stNormal

			case IAC:
				t.sb.WriteByte(IAC)
				t.state = stSB

			default:
				t.sb.WriteByte(IAC)
				t.sb.WriteByte(b)
				t.state = stSB

			}
		}
	}
	return 0, nil
}

func (t *TelnetReader) Prime(c net.Conn, timeout time.Duration) error {
	if t.pending != nil {
		return nil
	}

	c.SetReadDeadline(time.Now().Add(timeout))
	buf := []byte{0}
	n, err := t.Read(buf)

	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return nil
		}

		return err
	}

	if n == 1 {
		b := buf[0]
		t.pending = &b
	}

	return nil
}
