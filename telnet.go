package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"sync"
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

type TelnetConn struct {
	c   net.Conn
	Env map[string]string

	dataCh  chan byte
	errCh   chan error
	closeCh chan struct{}

	writeMu sync.Mutex
}

func NewTelnetConn(c net.Conn) *TelnetConn {
	t := &TelnetConn{
		Env: make(map[string]string),

		c:       c,
		dataCh:  make(chan byte, 4096),
		errCh:   make(chan error, 1),
		closeCh: make(chan struct{}),
	}

	go t.parseLoop()
	return t
}

func (t *TelnetConn) requestEnv() {
	var telnetEnvSend = []byte{
		IAC, SB, NEW_ENVIRON, SEND, IAC, SE,
	}
	t.writeRaw(telnetEnvSend)
}

func (t *TelnetConn) handleCmd(cmd, opt byte) {
	switch cmd {
	case DO:
		if opt == NEW_ENVIRON || opt == ECHO {
			t.WriteIAC(WILL, opt) // agree to send ENV
		} else {
			t.WriteIAC(WONT, opt)
		}
	case DONT:
		t.WriteIAC(WONT, opt)
	case WILL:
		if opt == NEW_ENVIRON {
			t.WriteIAC(DO, opt) // agree to receive ENV
			t.requestEnv()
		} else if opt == ECHO {
			t.WriteIAC(DO, opt)
		} else {
			t.WriteIAC(DONT, opt)
		}
	case WONT:
		t.WriteIAC(DONT, opt)
	}
}

func (t *TelnetConn) parseLoop() {
	r := bufio.NewReader(t.c)
	sb := new(bytes.Buffer)
	cmd := byte(0)
	const (
		stNormal = iota
		stIAC
		stIACIgnore
		stIACCmd
		stSB
		stSB_IAC
	)
	state := stNormal

	for {
		b, err := r.ReadByte()
		if err != nil {
			select {
			case t.errCh <- err:
			default:
			}
			close(t.dataCh)
			return
		}

		switch state {
		case stNormal:
			// IAC telnet command
			if b == IAC {
				state = stIAC
				continue
			}
			select {
			case t.dataCh <- b:
			case <-t.closeCh:
				return
			}

		case stIAC:
			switch b {
			// IAC escaped
			case IAC:
				state = stNormal
				select {
				case t.dataCh <- IAC:
				case <-t.closeCh:
					return
				}

			// start recording in sb buffer
			case SB:
				sb.Reset()
				state = stSB

			case WILL, WONT, DO, DONT:
				cmd = b
				state = stIACCmd

			// ignore other then sb
			default:
				state = stNormal
			}

		case stIACIgnore:
			state = stNormal

		case stIACCmd:
			t.handleCmd(cmd, b)
			state = stNormal

		case stSB:
			if b == IAC {
				state = stSB_IAC
				continue
			}
			sb.WriteByte(b)

		case stSB_IAC:
			switch b {
			case SE:
				parseEnv(sb.Bytes(), t.Env)
				state = stNormal

			case IAC:
				sb.WriteByte(IAC)
				state = stSB

			default:
				sb.WriteByte(IAC)
				sb.WriteByte(b)
				state = stSB

			}
		}
	}

}

func (t *TelnetConn) Read(p []byte) (int, error) {
	i := 0
	for i < len(p) {
		select {
		case b, ok := <-t.dataCh:
			if !ok {
				select {
				case err := <-t.errCh:
					if i == 0 {
						return 0, err
					}
					return i, nil
				default:
					if i == 0 {
						return 0, io.EOF
					}
					return i, nil
				}
			}
			p[i] = b
			i++
			if i > 0 {
				return i, nil
			}
		case <-t.closeCh:
			return 0, io.EOF
		}
	}
	return i, nil
}

func (t *TelnetConn) writeRaw(p []byte) (int, error) {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	n, err := t.c.Write(p)
	return n, err
}

func (t *TelnetConn) Write(p []byte) (int, error) {
	escaped := make([]byte, 0, len(p)+4)
	for _, bb := range p {
		escaped = append(escaped, bb)
		if bb == IAC {
			escaped = append(escaped, bb)
		}
	}

	// TODO: calculate the correct size on error
	_, err := t.writeRaw(escaped)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (t *TelnetConn) WriteIAC(cmd, opt byte) error {
	_, err := t.writeRaw([]byte{IAC, cmd, opt})
	return err
}
