package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
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

	var wg sync.WaitGroup
	done := make(chan bool)
loop:
	for {
		select {
		case <-signals:
			break loop
		case session, ok := <-sessions:
			if ok {
				defer session.Session.Close()
				host := session.Host
				outScanner := initScanner(session)
				requestPty(session.Session)
				wg.Add(1)
				go runRemote(command, session, &wg)
				go output(host, outScanner)
			} else {
				go func() {
					wg.Wait()
					done <- true
				}()
				sessions = nil
			}
		case <-done:
			break loop
		}
	}
}

func remoteCommand() string {
	return strings.Join(flag.Args()[1:], " ")
}

func initScanner(session Session) *bufio.Scanner {
	reader, _ := session.Session.StdoutPipe()
	scanner := bufio.NewScanner(reader)
	return scanner
}

func runRemote(command string, session Session, wg *sync.WaitGroup) {
	err := session.Session.Run(command)
	if err != nil {
		errLogger.Print("remote command failed: " + err.Error())
	}
	wg.Done()
}

func output(host string, scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Println("[" + host + "] " + scanner.Text())
	}
}
