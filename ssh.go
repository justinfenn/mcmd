package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"github.com/howeyc/gopass"
)

type Session struct {
	Host    string
	Session *ssh.Session
}

func openSessions(hostConfig HostConfig) chan Session {
	sessions := make(chan Session)
	clientConfig := clientConfig(hostConfig.User, hostConfig.Auth)
	var wg sync.WaitGroup
	for _, host := range hostConfig.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			session := connect(host, clientConfig)
			sessions <- Session{host, session}
		}(host)
	}
	go func() {
		wg.Wait()
		close(sessions)
	}()
	return sessions
}

func clientConfig(user string, auth AuthConfig) *ssh.ClientConfig {
	var configuredMethod ssh.AuthMethod
	if auth.Password {
		configuredMethod = ssh.Password(getPassword())
	} else {
		configuredMethod = ssh.PublicKeys(getKeySigner(auth.Privatekey))
	}

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			configuredMethod,
		},
	}
	return &config
}

func connect(host string, config *ssh.ClientConfig) *ssh.Session {
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		errLogger.Print(err)
		return nil
	}
	session, err := conn.NewSession()
	if err != nil {
		errLogger.Print(err)
		return nil
	}
	return session
}

func getPassword() string {
	fmt.Print("Password: ")
	return string(gopass.GetPasswd())
}

func getKeySigner(filename string) ssh.Signer {
	filename = os.ExpandEnv(filename)
	keyContents, err := ioutil.ReadFile(filename)
	if err != nil {
		errLogger.Fatalln(err)
	}
	signer, err := ssh.ParsePrivateKey(keyContents)
	if err != nil {
		errLogger.Fatalln(err)
	}
	return signer
}
