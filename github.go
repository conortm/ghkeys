package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type tokenSource struct {
	token *oauth2.Token
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}

// Github holds github token and client
type Github struct {
	token  string
	client *github.Client
}

func newGithub(token string) Github {
	// TODO: https://github.com/sourcegraph/apiproxy
	client := github.NewClient(oauth2.NewClient(oauth2.NoContext, &tokenSource{
		&oauth2.Token{
			AccessToken: token,
			TokenType:   "token",
		},
	}))

	//* DEBUG: print rate-limit.
	fmt.Println("<DEBUG>")
	rate, _, err := client.RateLimit()
	if err != nil {
		fmt.Printf("Error fetching GitHub API Rate Limit: %#v\n", err)
	} else {
		fmt.Printf("GitHub API Rate Limit: %#v\n", rate)
	}
	fmt.Println("</DEBUG>")
	// END DEBUG. */

	return Github{client: client, token: token}
}

func (g *Github) getTeamID(orgName string) (int, error) {
	orgNameArray := strings.Split(orgName, "/")
	org := orgNameArray[0]
	name := orgNameArray[1]
	teams, _, err := g.client.Organizations.ListTeams(org, nil)
	if err != nil {
		return 0, err
	}
	for _, team := range teams {
		if strings.EqualFold(*team.Name, name) {
			return *team.ID, nil
		}
	}
	return 0, errors.New("Team ID not found.")
}

var githubTeamMembers = map[string][]string{}

func (g *Github) getMembersOfTeam(orgName string) ([]string, error) {
	members, ok := githubTeamMembers[orgName]
	if ok {
		return members, nil
	}
	teamID, err := g.getTeamID(orgName)
	if err != nil {
		return members, err
	}
	githubUsers, _, err := g.client.Organizations.ListTeamMembers(teamID, nil)
	if err != nil {
		return members, err
	}
	for _, githubUser := range githubUsers {
		members = append(members, *githubUser.Login)
	}
	githubTeamMembers[orgName] = members
	return members, nil
}

var githubUserKeys = map[string][]string{}

func (g *Github) getKeysOfUser(user string) ([]string, error) {
	keys, ok := githubUserKeys[user]
	if ok {
		return keys, nil
	}
	githubKeys, _, err := g.client.Users.ListKeys(user, nil)
	if err != nil {
		return keys, err
	}
	for _, githubKey := range githubKeys {
		keys = append(keys, *githubKey.Key)
	}
	githubUserKeys[user] = keys
	return keys, nil
}

func appendUserIfMissing(users []string, newUser string) []string {
	for _, user := range users {
		if user == newUser {
			return users
		}
	}
	return append(users, newUser)
}

func (g *Github) getKeys(users, teams []string) []string {
	var keys []string
	// add members of teams to array of users
	teamMembersChan := make(chan []string)
	for _, orgName := range teams {
		go func(orgName string) {
			members, _ := g.getMembersOfTeam(orgName)
			// TODO: handle error
			teamMembersChan <- members
		}(orgName)
	}
	for i := 0; i < len(teams); i++ {
		members := <-teamMembersChan
		for _, member := range members {
			users = appendUserIfMissing(users, member)
		}
	}
	// get keys of each user
	userKeysChan := make(chan []string)
	for _, user := range users {
		go func(user string) {
			userKeys, _ := g.getKeysOfUser(user)
			// TODO: handle error
			userKeysChan <- userKeys
		}(user)
	}
	for i := 0; i < len(users); i++ {
		userKeys := <-userKeysChan
		keys = append(keys, userKeys...)
	}
	return keys
}
