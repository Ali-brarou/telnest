package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	Addr string
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Println("telnest listening on", s.Addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handle(c)
	}
}

func checkCVE_2026_24061(tel *TelnetConn) bool {
	_, ok := tel.Env["USER"]
	return ok
}

func (s *Server) handle(c net.Conn) {
	defer func() {
		log.Printf("[%s] disconnected\n", c.RemoteAddr())
		c.Close()
	}()

	log.Printf("[%s] connected", c.RemoteAddr())

	tel := NewTelnetConn(c)
	tel.WriteIAC(DO, NEW_ENVIRON)
	/* sleep a little bit to detect early negotiations */
	time.Sleep(300 * time.Millisecond)

	/* if the attacker put a USER env then skip auth */
	authed := false
	if checkCVE_2026_24061(tel) {
		authed = true
		log.Printf(
			"[%s] TELNET CVE-2026-24061 exploited",
			c.RemoteAddr(),
		)
	}

	rw := bufio.NewReadWriter(
		bufio.NewReader(tel),
		bufio.NewWriter(tel),
	)

	// send the banner
	rw.WriteString("\n\rLinux 6.17.9-arch1-1 (ab-device) (pts/5)\n\r\n\r")
	rw.Flush()

	user := "root"
	for !authed {
		// send login prompt
		rw.WriteString("ab-device login: ")
		rw.Flush()
		c.SetDeadline(time.Now().Add(60 * time.Second))
		user, err := rw.ReadString('\n')
		user = strings.TrimSpace(user)
		if err != nil {
			return
		}

		if checkCVE_2026_24061(tel) {
			log.Printf(
				"[%s] TELNET CVE-2026-24061 exploited",
				c.RemoteAddr(),
			)
			break
		}

		// hide password client side
		tel.WriteIAC(WILL, ECHO)
		// send password prompt
		rw.WriteString("Password : ")
		rw.Flush()
		c.SetDeadline(time.Now().Add(60 * time.Second))
		pass, err := rw.ReadString('\n')
		pass = strings.TrimSpace(pass)
		if err != nil {
			return
		}
		// show text client side
		tel.WriteIAC(WONT, ECHO)
		rw.WriteString("\n")
		rw.Flush()

		log.Printf(
			"[%s] TELNET auth attempt user=%q pass=%q env=%v",
			c.RemoteAddr(),
			user,
			pass,
			tel.Env,
		)

		if checkCVE_2026_24061(tel) {
			log.Printf(
				"[%s] TELNET CVE-2026-24061 exploited",
				c.RemoteAddr(),
			)
			break
		}

		if user == "root" && pass == "1234" {
			break
		} else {
			time.Sleep(2 * time.Second)
			rw.WriteString("\r\nLogin incorrect\r\n")
			rw.Flush()
		}
	}

	log.Printf("[%s] got a shell\n", c.RemoteAddr())

	prompt := fmt.Sprintf("[%s@ab-device ~]# ", user)
	for {
		rw.WriteString(prompt)
		rw.Flush()

		c.SetDeadline(time.Now().Add(5 * time.Minute))
		cmd, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		log.Printf("[%s] Shell command %q\n", c.RemoteAddr(), strings.TrimSpace(cmd))

		rw.WriteString(fakeShell(cmd))
		rw.Flush()
	}
}
