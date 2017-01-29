package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yale-cpsc-213/servers/questions"
)

const usage string = `
Creates and grades server homework for CPSC213. Usage:
cpsc213servers test URL
`

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 3 {
		fmt.Println(usage)
		return
	}
	url := os.Args[2]
	switch os.Args[1] {
	case "test":
		questions.TestAll(url, true)
	default:
		fmt.Println("ERROR! Bad input. See below for usage.")
		fmt.Println(usage)
		return
	}

}
