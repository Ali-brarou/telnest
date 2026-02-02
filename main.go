package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	logFileName := "telnest.log"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer logFile.Close()

	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)
	fmt.Println("logging to file:", logFileName)

	// https://www.shodan.io/search?query=telnet
	ports := []string{
		":23",
		":2323",
		":4000",
	}

	for _, addr := range ports {
		server := NewServer(addr)

		go func(s *Server) {
			if err = server.ListenAndServe(); err != nil {
				log.Println("listen error:", err)
			}
		}(server)
	}

	select {}
}
