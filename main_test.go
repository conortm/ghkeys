package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	configFile, _ := ioutil.TempFile("", "ghkeys")
	defer os.RemoveAll(configFile.Name())

	configContent := `---
github_token: github_token_value
users:
  - username: user_1
    github_users:
      - github_user_1
      - github_user_2
  - username: user_2
    github_teams:
      - MyOrg/Team 1
      - MyOrg/Team 2
`
	ioutil.WriteFile(configFile.Name(), []byte(configContent), os.ModePerm)

	testConfig, err := newConfig(configFile.Name())
	assert.Nil(t, err)
	assert.Equal(t, "github_token_value", testConfig.GithubToken)
	assert.Len(t, testConfig.Users, 2)
	assert.Equal(t, "user_1", testConfig.Users[0].Username)
	assert.Len(t, testConfig.Users[0].GithubUsers, 2)
	assert.Equal(t, "github_user_1", testConfig.Users[0].GithubUsers[0])
	assert.Nil(t, testConfig.Users[0].GithubTeams)
	assert.Nil(t, testConfig.Users[1].GithubUsers)
	assert.Len(t, testConfig.Users[1].GithubTeams, 2)
	assert.Equal(t, "MyOrg/Team 1", testConfig.Users[1].GithubTeams[0])

}
