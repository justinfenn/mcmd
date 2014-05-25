package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	errLogger = log.New(os.Stderr, "", 0)
)

func main() {
	flag.Parse()
	config := loadConfig()

	command := remoteCommand()

	sessions := openSessions(config)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT)

	var commandsDone []chan bool
	var connectedSessions []Session
	for session := range sessions {
		if session.Session == nil {
			continue
		}
		defer session.CloseOnce()
		host := session.Host
		outScanner := initScanner(session)
		d := make(chan bool)
		commandsDone = append(commandsDone, d)
		connectedSessions = append(connectedSessions, session)
		requestPty(session.Session)
		go runRemote(command, session, d)
		go output(host, outScanner)
	}

	go func() {
		<-signals
		for _, session := range connectedSessions {
			session.CloseOnce()
		}
	}()

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
