package main

import (
	"bufio"
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

	command := remoteCommand()

	sessions := openSessions(config)

	var commandsDone []chan bool
	for session := range sessions {
		if session.Session == nil {
			continue
		}
		defer session.Session.Close()
		host := session.Host
		outScanner := initScanner(session)
		d := make(chan bool)
		commandsDone = append(commandsDone, d)
		go runRemote(command, session, d)
		go output(host, outScanner)
	}

	for _, d := range commandsDone {
		<-d
	}

	fmt.Println("done")
}

func remoteCommand() string {
	return strings.Join(flag.Args()[1:], " ")
}

func initScanner(session Session) *bufio.Scanner {
	reader, _ := session.Session.StdoutPipe()
	scanner := bufio.NewScanner(reader)
	return scanner
}

func runRemote(command string, session Session, done chan bool) {
	err := session.Session.Run(command)
	if err != nil {
		errLogger.Print("remote command failed: " + err.Error())
	}
	done <- true
}

func output(host string, scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Println("[" + host + "] " + scanner.Text())
	}
	fmt.Println("---")
}
