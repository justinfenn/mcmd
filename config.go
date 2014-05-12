package main

import (
	"flag"
	"io/ioutil"

	"gopkg.in/v1/yaml"
)

type HostConfig struct {
	User  string
	Hosts []string
	Auth  AuthConfig
}

type AuthConfig struct {
	Password   bool
	Privatekey string
}

func loadConfig() HostConfig {
	rawYaml := readHostfile()
	var result HostConfig
	err := yaml.Unmarshal(rawYaml, &result)
	if err != nil {
		errLogger.Fatalln("Error parsing hostfile:", err)
	}
	return result
}

func readHostfile() []byte {
	filename := flag.Arg(0)
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		errLogger.Fatalln(err)
	}
	return fileContents
}
