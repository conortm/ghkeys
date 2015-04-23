package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	GithubToken string `yaml:"github_token"`
	Users       []struct {
		Username    string   `yaml:"username"`
		GithubUsers []string `yaml:"github_users"`
		GithubTeams []string `yaml:"github_teams"`
	}
}

func newConfig(configFilename string) (config, error) {
	config := config{}
	data, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
