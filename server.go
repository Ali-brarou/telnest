package main

import (
	"bufio"
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

func (s *Server) handle(c net.Conn) {
	defer func() {
		log.Printf("[%s] disconnected\n", c.RemoteAddr())
		c.Close()
	}()

	c.SetDeadline(time.Now().Add(10 * time.Second))

	log.Printf("[%s] connected", c.RemoteAddr())
	/*
		buf := make([]byte, 4096)
		_, err := c.Read(buf)
		if err != nil {
			return
		}
		env := parseEnv(buf)
	*/

	rw := bufio.NewReadWriter(
		bufio.NewReader(c),
		bufio.NewWriter(c),
	)

	/* send the banner */
	rw.WriteString("\n\rLinux 6.17.9-arch1-1 (ab-device) (pts/5)\n\r\n\rab-device login: ")
	rw.Flush()
	user, err := rw.ReadString('\n')
	if err != nil {
		return
	}

	rw.Write([]byte{IAC, WILL, ECHO})
	rw.WriteString("Password : ")
	rw.Flush()
	pass, err := rw.ReadString('\n')
	if err != nil { 
		return
	}
	rw.Write([]byte{IAC, WONT, ECHO, 0xa})
	rw.Flush()

	log.Printf(
		"[%s] TELNET auth attempt user=%q pass=%q",
		c.RemoteAddr(),
		strings.TrimSpace(user),
		strings.TrimSpace(pass),
	)

/*	if string(user) == "root" && string(pass) == "root" { */
		for {
			rw.WriteString("[root@ab-device ~]# ")
			rw.Flush()

			cmd, err := rw.ReadString('\n')
			if err != nil {
			 	return	
			}
			log.Printf("[%s] Shell command %q\n", c.RemoteAddr(), strings.TrimSpace(cmd))

			rw.WriteString(fakeShell(cmd))
			rw.Flush()
		}
	/* } */ 

	time.Sleep(2 * time.Second)
	rw.WriteString("\r\nLogin incorrect\r\n")
	rw.Flush()
}
