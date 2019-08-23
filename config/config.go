package config

import (
	"fmt"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config struct
type Config struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Name             string `yaml:"name"`
	JSONPath         string `yaml:"jsonpath"`
	Regex            string `yaml:"regex"`
	AdmissionMessage string `yaml:"admissionMessage"`
}

// LoadConfig returns new Config
func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	configFile := filepath.Join(configPath, "config.yaml")
	return load(configFile)
}

func load(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		return &Config{}, err
	}
	return read(file)
}

func read(file io.Reader) (*Config, error) {
	c := &Config{}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		glog.Errorf("Error reading config file :: %s", err.Error())
		return c, err
	}

	if len(b) != 0 {
		err := yaml.Unmarshal(b, c)
		if err != nil {
			glog.Errorf("Error marshalling config :: %s", err.Error())
			return c, err
		}
	}
	return c, nil
}

func ValidateConfig(c *Config) error {

	if c == nil {
		glog.Error("Config object not initialized")
		return fmt.Errorf("Config Object not initialized")
	}
	if len(c.Rules) == 0 {
		glog.Error("No rules found")
		return fmt.Errorf("No rules found")
	}

	for _, rule := range c.Rules {
		if rule.Name == "" {
			glog.Error("Empty rule name")
			return fmt.Errorf("Empty rule name found")
		}
		if rule.JSONPath == "" {
			glog.Error("Empty jsonpath")
			return fmt.Errorf("Empty jsonpath")
		}
		if rule.AdmissionMessage == "" {
			glog.Error("Empty admission message")
			return fmt.Errorf("Empty admission message")
		}
		if rule.Regex == "" {
			glog.Error("Empty regex")
			return fmt.Errorf("Empty rule regex")
		}
	}
	return nil
}
