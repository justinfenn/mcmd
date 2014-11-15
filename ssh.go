package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Session struct {
	Host    string
	Session *ssh.Session
}

func openSessions(hostConfig HostConfig) chan Session {
	sessions := make(chan Session)
	var wg sync.WaitGroup
	for _, host := range hostConfig.Hosts {
		wg.Add(1)
		clientConfig := clientConfig(hostConfig.User, hostConfig.Privatekey)
		go func(host string) {
			defer wg.Done()
			session, err := connect(host, clientConfig)
			if err != nil {
				errLogger.Print(err.Error())
			} else {
				sessions <- Session{host, session}
			}
		}(host)
	}
	go func() {
		wg.Wait()
		close(sessions)
	}()
	return sessions
}

func clientConfig(user, pkFile string) *ssh.ClientConfig {
	signerCallback := func() ([]ssh.Signer, error) {
		return getAllSigners(pkFile)
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(signerCallback),
			ssh.PasswordCallback(getPassword),
		},
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

var sharedPassword string
var pwOnce sync.Once

func getPassword() (string, error) {
	pwOnce.Do(func() {
		fmt.Print("Password: ")
		sharedPassword = string(gopass.GetPasswd())
	})
	return sharedPassword, nil
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
