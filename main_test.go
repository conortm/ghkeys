package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	server *httptest.Server
	client *githubClient
)

func setup() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := "Not Found"
		switch r.URL.Path {
		case "/orgs/MyOrg/teams":
			response = `[{"id":1, "name":"Team 1"},{"id":2, "name":"Team 2"}]`
		case "/orgs/MyOtherOrg/teams":
			response = `[{"id":3, "name":"Team 3"}]`
		case "/teams/1/members":
			response = `[{"login":"github_user_1"},{"login":"github_user_2"}]`
		case "/teams/2/members":
			response = `[{"login":"github_user_3"}]`
		case "/teams/3/members":
			response = `[{"login":"github_user_4"}]`
		case "/users/github_user_1/keys":
			response = `[{"key":"github_user_1_key_1"},{"key":"github_user_1_key_2"}]`
		case "/users/github_user_2/keys":
			response = `[{"key":"github_user_2_key_1"}]`
		case "/users/github_user_3/keys":
			response = `[{"key":"github_user_3_key_1"}]`
		case "/users/github_user_4/keys":
			response = `[{"key":"github_user_4_key_1"}]`
		}
		fmt.Fprint(w, response)
	}))
	client = newGithubClient("token")
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func TestConfig(t *testing.T) {
	testConfig, err := newConfig("config.example.yml")

	assert.Nil(t, err)
	assert.Equal(t, "my_github_token", testConfig.GithubToken)
	assert.Len(t, testConfig.Users, 2)
	assert.Equal(t, "superadmin", testConfig.Users[0].Username)
	assert.Len(t, testConfig.Users[0].GithubUsers, 1)
	assert.Equal(t, "github_user_1", testConfig.Users[0].GithubUsers[0])
	assert.Len(t, testConfig.Users[0].GithubTeams, 2)
	assert.Equal(t, "MyOrg/Team 1", testConfig.Users[0].GithubTeams[0])
}

func TestGetTeamID(t *testing.T) {
	setup()
	defer teardown()

	teamID, err := client.getTeamID("MyOrg/Team 2")

	assert.Nil(t, err)
	assert.Equal(t, 2, teamID)
}

func TestGetMembersOfTeam(t *testing.T) {
	setup()
	defer teardown()

	membersOfTeam, err := client.getMembersOfTeam("MyOrg/Team 2")

	assert.Nil(t, err)
	assert.Len(t, membersOfTeam, 1)
	assert.Equal(t, "github_user_3", membersOfTeam[0])
}

func TestGetKeysOfUser(t *testing.T) {
	setup()
	defer teardown()

	keysOfUser, err := client.getKeysOfUser("github_user_1")

	assert.Nil(t, err)
	assert.Len(t, keysOfUser, 2)
	assert.Equal(t, "github_user_1_key_2", keysOfUser[1])
}

func TestGetKeys(t *testing.T) {
	setup()
	defer teardown()

	users := []string{"github_user_1"}
	teams := []string{"MyOrg/Team 1", "MyOrg/Team 2"}
	expectedKeys := []string{"github_user_1_key_1", "github_user_1_key_2", "github_user_2_key_1", "github_user_3_key_1"}

	keys := client.getKeys(users, teams)

	// TODO: Better way to do this?
	assert.Len(t, keys, len(expectedKeys))
	for _, expectedKey := range expectedKeys {
		assert.Contains(t, keys, expectedKey)
	}
}
