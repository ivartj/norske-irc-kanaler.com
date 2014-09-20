package main

import (
	yaml "gopkg.in/yaml.v1"
	"os"
	"io/ioutil"
	"fmt"
)

type config struct {
	Method string `yaml:"method"`
	FastcgiPath string `yaml:"fastcgi-path"`
	HttpPort uint `yaml:"http-port"`
	DatabasePath string `yaml:"database-path"`
	AssetsPath string `yaml:"assets-path"`
	LogPath string `yaml:"log-path"`
}

func confParse(filename string) *config {
	conf := &config{}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file '%s': %s\n", filename, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on reading from file '%s': %s.\n", filename, err.Error())
		os.Exit(1)
	}

	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error on parsing configuration file '%s': %s.\n", filename, err.Error())
		os.Exit(1)
	}

	confSanityCheck(conf)
	return conf
}

func confSanityCheck(conf *config) {
	switch conf.Method {
	case "":
		fmt.Fprintln(os.Stderr, "Configuration error: Missing 'method' field value. Specify either 'http' or 'fastcgi'.")
		os.Exit(1)
	case "http":
		if conf.HttpPort == 0 || conf.HttpPort > 65545 {
			fmt.Fprintf(os.Stderr, "Configuration error: Invalid HTTP port number '%d'.\n", conf.HttpPort)
			os.Exit(1)
		}
	case "fastcgi":
		if conf.FastcgiPath == "" {
			fmt.Fprintln(os.Stderr, "Configuration error: Missing 'fastcgi-path' field value.")
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Configuration error: Unrecognized 'method' field value.")
		os.Exit(1)
	}

	if conf.DatabasePath == "" {
		fmt.Fprintln(os.Stderr, "Configuration error: Missing 'database-path' field value.")
		os.Exit(1)
	}

	if conf.AssetsPath == "" {
		fmt.Fprintf(os.Stderr, "Configuration error: Missing 'assets-path' field value.")
		os.Exit(1)
	}
}
