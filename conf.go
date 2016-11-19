package main

import (
	yaml "gopkg.in/yaml.v1"
	"os"
	"io/ioutil"
	"fmt"
)

type confSet struct {
	WebsiteTitle string `yaml:"website-title"`
	WebsiteDescription string `yaml:"website-description"`
	ServeMethod string `yaml:"serve-method"`
	FastcgiPath string `yaml:"fastcgi-path"`
	HttpPort uint `yaml:"http-port"`
	DatabasePath string `yaml:"database-path"`
	AssetsPath string `yaml:"assets-path"`
	LogPath string `yaml:"log-path"`
	Approval bool `yaml:"approval"`
	Password string `yaml:"password"`
	ReloadTemplate bool `yaml:"reload-templates"`
	IrssiLogsPath string `yaml:"irssi-logs-path"`
}

type conf struct {
	set confSet
}

func (cfg *conf) WebsiteTitle() string { return cfg.set.WebsiteTitle }
func (cfg *conf) WebsiteDescription() string { return cfg.set.WebsiteDescription }
func (cfg *conf) ServeMethod() string { return cfg.set.ServeMethod }
func (cfg *conf) FastcgiPath() string { return cfg.set.FastcgiPath }
func (cfg *conf) HttpPort() uint { return cfg.set.HttpPort }
func (cfg *conf) DatabasePath() string { return cfg.set.DatabasePath }
func (cfg *conf) AssetsPath() string { return cfg.set.AssetsPath }
func (cfg *conf) LogPath() string { return cfg.set.LogPath }
func (cfg *conf) Approval() bool { return cfg.set.Approval }
func (cfg *conf) Password() string { return cfg.set.Password }
func (cfg *conf) ReloadTemplate() bool { return cfg.set.ReloadTemplate }
func (cfg *conf) IrssiLogsPath() string { return cfg.set.IrssiLogsPath }

func confNew() *conf {
	return &conf{}
}

func (cfg *conf) ParseFile(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Failed to open file '%s': %s", filename, err.Error())
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Error on reading from file '%s': %s", filename, err.Error())
	}

	err = yaml.Unmarshal(bytes, &(cfg.set))
	if err != nil {
		return fmt.Errorf("Error on parsing configuration file '%s': %s", filename, err.Error())
	}

	err = cfg.Validate()
	if err != nil {
		return fmt.Errorf("Failed to validate configuration: %s", err.Error())
	}

	return nil
}

func (cfg *conf) Validate() error {
	switch cfg.ServeMethod() {
	case "":
		return fmt.Errorf("Configuration error: Missing 'method' field value. Specify either 'http' or 'fastcgi'.\n")
	case "http":
		if cfg.HttpPort() == 0 || cfg.HttpPort() > 65545 {
			return fmt.Errorf("Configuration error: Invalid HTTP port number '%d'.\n", cfg.HttpPort())
		}
	case "fastcgi":
		if cfg.FastcgiPath() == "" {
			return fmt.Errorf("Configuration error: Missing 'fastcgi-path' field value.\n")
		}
	default:
		return fmt.Errorf("Configuration error: Unrecognized 'method' field value.\n")
	}

	if cfg.DatabasePath() == "" {
		return fmt.Errorf("Configuration error: Missing 'database-path' field value.\n")
	}

	if cfg.AssetsPath() == "" {
		return fmt.Errorf("Configuration error: Missing 'assets-path' field value.\n")
	}

	if cfg.Password() == "" {
		return fmt.Errorf("Configuration error: Missing 'password' field value.\n")
	}

	if cfg.WebsiteTitle() == "" {
		return fmt.Errorf("Configuration error: Missing 'website-title' field value.\n")
	}

	if cfg.WebsiteDescription() == "" {
		return fmt.Errorf("Configuration error: Missing 'website-description' field value.\n")
	}

	return nil
}

