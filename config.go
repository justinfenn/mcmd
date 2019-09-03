package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v1"
)

const (
	CONFIG_SUBDIR            = "mcmd"
	DEFAULT_KNOWN_HOSTS_FILE = "$HOME/.ssh/known_hosts"
)

type HostConfig struct {
	User       string
	Privatekey string
	Hosts      []string
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
	configFileParam := flag.Arg(0)
	for _, file := range getConfigLocationCandidates(configFileParam) {
		if exists(file) {
			return getContents(file)
		}
	}
	errLogger.Fatalf("hostfile %s not found", configFileParam)
	return nil
}

func getConfigLocationCandidates(configFileParam string) []string {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		errLogger.Fatalf("no user config dir found")
	}
	configDir := path.Join(configRoot, CONFIG_SUBDIR)
	return []string{configFileParam,
		path.Join(configDir, configFileParam+".yml"),
		path.Join(configDir, configFileParam+".yaml")}
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func getContents(filename string) []byte {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		errLogger.Fatalln(err)
	}
	return fileContents
}
