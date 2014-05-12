package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	errLogger = log.New(os.Stderr, "", 0)
)

func main() {
	flag.Parse()
	config := loadConfig()

	sessions := openSessions(config)
	defer func() {
		for _, session := range sessions {
			session.Close()
		}
	}()

	command := remoteCommand()
	for host, session := range sessions {
		out, err := session.CombinedOutput(command)
		if err != nil {
			errLogger.Print("remote command failed: " + err.Error())
		}
		fmt.Println("[" + host + "] " + string(out))
	}
	fmt.Println("done")
}

func remoteCommand() string {
	return strings.Join(flag.Args()[1:], " ")
}
