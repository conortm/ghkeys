package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testdataConfigFilename     = "testdata/config.test.yml"
	testAuthorizedKeysFilename = "testdata/authorized_keys"
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
	// Test that test config file jives with config struct
	testConfig, err := newConfig(testdataConfigFilename)
	assert.Nil(t, err)
	assert.Equal(t, "my_github_token", testConfig.GithubToken)
	assert.Len(t, testConfig.Users, 2)
	assert.Equal(t, "superadmin", testConfig.Users[0].Username)
	assert.Len(t, testConfig.Users[0].GithubUsers, 1)
	assert.Equal(t, "github_user_1", testConfig.Users[0].GithubUsers[0])
	assert.Len(t, testConfig.Users[0].GithubTeams, 2)
	assert.Equal(t, "MyOrg/Team 1", testConfig.Users[0].GithubTeams[0])

	// Test nonexistent config file error
	_, err = newConfig("testdata/thisfileshouldnotexist.yml")
	assert.Error(t, err)

	// Test malformed yaml error
	configFile, _ := ioutil.TempFile("", "ghkeys-config")
	defer os.RemoveAll(configFile.Name())
	configContent := "invalid config yml"
	ioutil.WriteFile(configFile.Name(), []byte(configContent), os.ModePerm)
	_, err = newConfig(configFile.Name())
	assert.Error(t, err)
}

func TestGetTeamID(t *testing.T) {
	setup()
	defer teardown()

	// Test Team ID that exists
	teamID, err := client.getTeamID("MyOrg/Team 2")
	assert.Nil(t, err)
	assert.Equal(t, 2, teamID)

	// Test invalid Team ID
	_, err = client.getTeamID("Invalid Team Name")
	assert.Error(t, err)

	// Test Team ID that does not exist
	_, err = client.getTeamID("MyOrg/Nonexistant Team")
	assert.Error(t, err)
}

func TestGetMembersOfTeam(t *testing.T) {
	setup()
	defer teardown()

	// Test getting members of valid Team
	membersOfTeam, err := client.getMembersOfTeam("MyOrg/Team 2")
	assert.Nil(t, err)
	assert.Len(t, membersOfTeam, 1)
	assert.Equal(t, "github_user_3", membersOfTeam[0])

	// Test getting members of invalid Team
	_, err = client.getMembersOfTeam("MyOrg/Invalid Team")
	assert.Error(t, err)
}

func TestGetKeysOfUser(t *testing.T) {
	setup()
	defer teardown()

	keysOfUser, err := client.getKeysOfUser("github_user_1")
	assert.Nil(t, err)
	assert.Len(t, keysOfUser, 2)
	assert.Equal(t, "github_user_1_key_2", keysOfUser[1])
}

func TestGetKeysOfUsersAndTeams(t *testing.T) {
	setup()
	defer teardown()

	users := []string{"github_user_1"}
	teams := []string{"MyOrg/Team 1", "MyOrg/Team 2"}
	expectedKeys := []string{"github_user_1_key_1", "github_user_1_key_2", "github_user_2_key_1", "github_user_3_key_1"}
	actualKeys := client.getKeysOfUsersAndTeams(users, teams)
	assert.Len(t, actualKeys, len(expectedKeys))
	for _, expectedKey := range expectedKeys {
		assert.Contains(t, actualKeys, expectedKey)
	}
}

func testUsernamesKeys(t *testing.T, expectedUsernamesKeys, actualUsernamesKeys map[string][]string) {
	assert.Len(t, actualUsernamesKeys, len(expectedUsernamesKeys))
	for expectedUsername, expectedKeys := range expectedUsernamesKeys {
		actualKeys, ok := actualUsernamesKeys[expectedUsername]
		assert.True(t, ok)
		for _, expectedKey := range expectedKeys {
			assert.Contains(t, actualKeys, expectedKey)
		}
	}
}

func TestGetUsernamesKeys(t *testing.T) {
	setup()
	defer teardown()

	config, err := newConfig(testdataConfigFilename)
	assert.Nil(t, err)

	// Test for single username
	expectedUsernamesKeys := map[string][]string{
		"superadmin": []string{
			"github_user_1_key_1",
			"github_user_1_key_2",
			"github_user_2_key_1",
			"github_user_3_key_1",
		},
	}
	actualUsernamesKeys := getUsernamesKeys(config, client, "superadmin")
	testUsernamesKeys(t, expectedUsernamesKeys, actualUsernamesKeys)

	// Test for all usernames
	expectedUsernamesKeys["admin"] = []string{
		"github_user_1_key_1",
		"github_user_1_key_2",
		"github_user_2_key_1",
		"github_user_4_key_1",
	}
	actualUsernamesKeys = getUsernamesKeys(config, client, "")
	testUsernamesKeys(t, expectedUsernamesKeys, actualUsernamesKeys)
}

func TestGetKeysOutput(t *testing.T) {
	expected := `github_user_1_key_1
github_user_1_key_2
github_user_1_key_3`
	keys := []string{"github_user_1_key_1", "github_user_1_key_2", "github_user_1_key_3"}
	actual := getKeysOutput(keys)
	assert.Equal(t, expected, actual)
}

func TestWriteKeysToUserAuthorizedKeysFile(t *testing.T) {
	// Test writing test keys to a test file
	keys := []string{"github_user_1_key_1", "github_user_1_key_2", "github_user_1_key_3"}
	userHomeDir, _ := ioutil.TempDir("", "ghkeys-userHomeDir")
	defer os.RemoveAll(userHomeDir)

	// No .ssh directory exists, we expect an error
	err := writeKeysToUserAuthorizedKeysFile(keys, userHomeDir)
	assert.Error(t, err)

	// Create a temp authorized_keys file with junk data
	authorizedKeysFilename := getAuthorizedKeysFilename(userHomeDir)
	os.MkdirAll(filepath.Dir(authorizedKeysFilename), os.ModePerm)
	authorizedKeysFile, _ := os.Create(authorizedKeysFilename)
	authorizedKeysFile.WriteString("junk")

	// Test writing to an authorized_keys file
	err = writeKeysToUserAuthorizedKeysFile(keys, userHomeDir)
	assert.Nil(t, err)

	// Compare test authorized_keys file to what we expect it to have
	expectedKeysFileContent, _ := ioutil.ReadFile(testAuthorizedKeysFilename)
	actualKeysFileContent, _ := ioutil.ReadFile(authorizedKeysFilename)
	assert.Equal(t, expectedKeysFileContent, actualKeysFileContent)
}
