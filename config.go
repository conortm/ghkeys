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
	c := config{}
	d, err := ioutil.ReadFile(configFilename)
	if err == nil {
		err = yaml.Unmarshal(d, &c)
	}
	return c, err
}
