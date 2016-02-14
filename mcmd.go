package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
)

var (
	errLogger = log.New(os.Stderr, "", 0)
	COLORS    = []func(...interface{}) string{
		color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgYellow).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
		color.New(color.FgWhite).SprintFunc(),
		color.New(color.FgCyan).SprintFunc(),
		color.New(color.FgBlack).SprintFunc(),
	}
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

func formatOutput(hostIndex int, host, str string) string {
	colorFunc := COLORS[hostIndex%len(COLORS)]
	return colorFunc("["+host+"] ") + str
}
