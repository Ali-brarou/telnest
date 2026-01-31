package main

import "fmt"

func main() {
	Server := NewServer(":2323")
	err := Server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
