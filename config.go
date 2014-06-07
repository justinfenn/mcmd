package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/v1/yaml"
)

const (
	XDG_CONFIG_HOME     = "XDG_CONFIG_HOME"
	DEFAULT_CONFIG_HOME = "$HOME/.config"
	CONFIG_SUBDIR       = "mcmd"
)

type HostConfig struct {
	User  string
	Hosts []string
	Auth  AuthConfig
}

type AuthConfig struct {
	Password   bool
	Privatekey string
	Agent      bool
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
	var configRoot, configDir string
	xdgHome := os.Getenv(XDG_CONFIG_HOME)
	if xdgHome != "" {
		configRoot = xdgHome
	} else {
		configRoot = os.ExpandEnv(DEFAULT_CONFIG_HOME)
	}
	configDir = path.Join(configRoot, CONFIG_SUBDIR)
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
