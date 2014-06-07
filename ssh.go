package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/howeyc/gopass"
)

type Session struct {
	Host    string
	Session *ssh.Session
	once    *sync.Once
}

func (s Session) CloseOnce() {
	s.once.Do(func() { s.Session.Close() })
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
			sessions <- Session{host, session, new(sync.Once)}
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
	} else if auth.Agent {
		configuredMethod = ssh.PublicKeys(getSignersFromAgent()...)
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

func getSignersFromAgent() []ssh.Signer {
	agent := getAgent()
	signers, err := agent.Signers()
	if err != nil {
		errLogger.Fatalln(err)
	}
	return signers
}

func getAgent() agent.Agent {
	agentSocket := os.Getenv("SSH_AUTH_SOCK")
	if agentSocket == "" {
		errLogger.Fatalln("No ssh agent found")
	}
	addr, err := net.ResolveUnixAddr("unix", agentSocket)
	if err != nil {
		errLogger.Fatalln(err)
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		errLogger.Fatalln(err)
	}
	agent := agent.NewClient(conn)
	if err != nil {
		errLogger.Fatalln(err)
	}
	return agent
}

func requestPty(session *ssh.Session) {
	err := session.RequestPty("xterm", 80, 40, nil)
	if err != nil {
		errLogger.Println(err)
	}
}
