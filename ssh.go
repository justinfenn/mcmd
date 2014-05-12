package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"github.com/howeyc/gopass"
)

func openSessions(hostConfig HostConfig) map[string]*ssh.Session {
	sessions := make(map[string]*ssh.Session)
	clientConfig := clientConfig(hostConfig.User, hostConfig.Auth)
	for _, host := range hostConfig.Hosts {
		conn, err := ssh.Dial("tcp", host, clientConfig)
		if err != nil {
			errLogger.Print(err)
			continue
		}
		session, err := conn.NewSession()
		if err != nil {
			errLogger.Print(err)
			continue
		}
		sessions[host] = session
	}
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
