package main

import (
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

	done := runCommands(command, config)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT)

	select {
	case <-signals:
		fmt.Println() // clean up for the next prompt
	case <-done:
	}
}

func remoteCommand() string {
	return strings.Join(flag.Args()[1:], " ")
}

func prependHost(host, str string) string {
	return "[" + host + "] " + str
}
