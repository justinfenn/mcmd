package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

func runCommands(command string, hostConfig HostConfig) chan bool {
	var wg sync.WaitGroup
	hostKeyCallback := createHostKeyCallback()
	for hostIndex, _ := range hostConfig.Hosts {
		wg.Add(1)
		go func(hostIndex int) {
			runCommand(command, hostIndex, hostConfig, hostKeyCallback)
			wg.Done()
		}(hostIndex)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	return done
}

func runCommand(command string, hostIndex int, hostConfig HostConfig, hostKeyCallback ssh.HostKeyCallback) {
	host := hostConfig.Hosts[hostIndex]
	clientConfig := clientConfig(hostConfig.User, hostConfig.Privatekey, hostKeyCallback)
	session, err := connect(host, clientConfig)
	if err != nil {
		errLogger.Print(err.Error())
		return
	}
	defer session.Close()
	scanner, err := prepareOutput(session)
	if err != nil {
		errLogger.Print(formatOutput(hostIndex, host, err.Error()))
		return
	}
	err = session.Start(command)
	if err != nil {
		errLogger.Print(formatOutput(hostIndex, host, err.Error()))
		return
	}
	for scanner.Scan() {
		fmt.Println(formatOutput(hostIndex, host, scanner.Text()))
	}
	err = scanner.Err()
	if err != nil {
		errLogger.Print(formatOutput(hostIndex, host, err.Error()))
	}
	err = session.Wait()
	if err != nil {
		errLogger.Print(formatOutput(hostIndex, host, err.Error()))
	}
}

func clientConfig(user, pkFile string, hostKeyCallback ssh.HostKeyCallback) *ssh.ClientConfig {
	signerCallback := func() ([]ssh.Signer, error) {
		return getAllSigners(pkFile)
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(signerCallback),
			ssh.PasswordCallback(getPassword),
		},
		HostKeyCallback: hostKeyCallback,
	}
	return &config
}

func connect(host string, config *ssh.ClientConfig) (*ssh.Session, error) {
	conn, err := ssh.Dial("tcp", completeAddress(host), config)
	if err != nil {
		return nil, err
	}
	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func prepareOutput(session *ssh.Session) (*bufio.Scanner, error) {
	err := requestPty(session)
	if err != nil {
		return nil, err
	}
	reader, _ := session.StdoutPipe()
	scanner := bufio.NewScanner(reader)
	return scanner, nil
}

var sharedPassword []byte
var sharedPasswordErr error
var pwOnce sync.Once

func getPassword() (string, error) {
	pwOnce.Do(func() {
		fmt.Print("Password: ")
		sharedPassword, sharedPasswordErr = gopass.GetPasswd()
	})
	return string(sharedPassword), sharedPasswordErr
}

func getAllSigners(pkFile string) ([]ssh.Signer, error) {
	var signers []ssh.Signer
	agentSigners, err := getSignersFromAgent()
	if err == nil {
		signers = append(signers, agentSigners...)
	}
	if pkFile != "" {
		keySigner, err := getKeySigner(pkFile)
		if err != nil {
			return nil, err
		}
		signers = append(signers, keySigner)
	}
	return signers, nil
}

func getKeySigner(filename string) (ssh.Signer, error) {
	filename = os.ExpandEnv(filename)
	keyContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(keyContents)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func getSignersFromAgent() ([]ssh.Signer, error) {
	agent, err := getAgent()
	if err != nil {
		return nil, err
	}
	signers, err := agent.Signers()
	if err != nil {
		return nil, err
	}
	return signers, nil
}

func getAgent() (agent.Agent, error) {
	agentSocket := os.Getenv("SSH_AUTH_SOCK")
	if agentSocket == "" {
		return nil, errors.New("no ssh agent found")
	}
	addr, err := net.ResolveUnixAddr("unix", agentSocket)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, err
	}
	agent := agent.NewClient(conn)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func requestPty(session *ssh.Session) error {
	err := session.RequestPty("xterm", 80, 40, nil)
	return err
}

func completeAddress(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = "ssh"
	}
	return net.JoinHostPort(host, port)
}

func createHostKeyCallback() ssh.HostKeyCallback {
	hostKeyCallback, err := knownhosts.New(os.ExpandEnv(DEFAULT_KNOWN_HOSTS_FILE))
	if err != nil {
		errLogger.Fatalln(err.Error())
	}
	return hostKeyCallback
}
