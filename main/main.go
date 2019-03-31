package main

import (
	"fmt"
	"nttp"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Too few arguments!")
		return
	}
	switch args[0] {
	case "help":
		fmt.Println("NTTP - Number Theoretic Transform Proxy")
		fmt.Println("Written by Chengyuan Ma")
		fmt.Println("Options: ")
		fmt.Println("	help: print what you see here")
		fmt.Println("	client [local] [server]: listen local and send to proxy")
		fmt.Println("	server [listen]: listen for relay connection and do proxy")
	case "client":
		if len(args) != 3 {
			fmt.Println("Too few or too many arguments!")
			return
		}
		nttp.ListenAsClient(args[1], args[2])
	case "server":
		if len(args) != 2 {
			fmt.Println("Too few or too many arguments!")
			return
		}
		nttp.ListenAsServer(args[1])
	}
}
