package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type YamlConfig struct {
	GlobalConfig `yaml:",inline"`

	Watcher WatcherConfig `yaml:"watcher"`
	API     ApiConfig     `yaml:"api"`
	SMTP    SMTPConfig    `yaml:"smtp"`
}

func (y *YamlConfig) Init() error {
	return nil
}

func (y *YamlConfig) setDefault() {
	y.GlobalConfig.setDefault()
	y.Watcher.setDefault()
	y.API.setDefault()
	y.SMTP.setDefault()
}

func (y *YamlConfig) check() (err ConfigError) {
	err = y.Watcher.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.GlobalConfig.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.API.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.SMTP.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}

func (y *YamlConfig) parser(filepath string) ParserError {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return NewParserError(err, err.Error())
	}

	err = yaml.Unmarshal(file, y)
	if err != nil {
		return NewParserError(err, err.Error())
	}

	return nil
}
