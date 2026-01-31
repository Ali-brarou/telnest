package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	logFile, err := os.OpenFile("telnet.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer logFile.Close()

	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)

	Server := NewServer(":2323")
	err = Server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
