package main

import (
	"errors"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type githubClient struct {
	*github.Client
}

func newGithubClient(token string) *githubClient {
	client := github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))
	return &githubClient{client}
}

func (gc *githubClient) getTeamID(orgName string) (int, error) {
	orgNameArray := strings.Split(orgName, "/")
	if len(orgNameArray) != 2 {
		return 0, errors.New("Team name must be in the form: MyOrg/Team Name")
	}
	org := orgNameArray[0]
	name := orgNameArray[1]
	opt := &github.ListOptions{PerPage: 25}
	for {
		teams, resp, err := gc.Organizations.ListTeams(org, opt)
		if err != nil {
			return 0, err
		}
		for _, team := range teams {
			if strings.EqualFold(*team.Name, name) {
				return *team.ID, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return 0, errors.New("Team ID not found.")
}

var githubTeamMembers = map[string][]string{}

func (gc *githubClient) getMembersOfTeam(orgName string) ([]string, error) {
	members, ok := githubTeamMembers[orgName]
	if ok {
		return members, nil
	}
	teamID, err := gc.getTeamID(orgName)
	if err != nil {
		return members, err
	}
	opt := &github.OrganizationListTeamMembersOptions{ListOptions: github.ListOptions{PerPage: 25}}
	for {
		githubUsers, resp, err := gc.Organizations.ListTeamMembers(teamID, opt)
		if err != nil {
			return members, err
		}
		for _, githubUser := range githubUsers {
			members = append(members, *githubUser.Login)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	githubTeamMembers[orgName] = members
	return members, nil
}

var githubUserKeys = map[string][]string{}

func (gc *githubClient) getKeysOfUser(user string) ([]string, error) {
	keys, ok := githubUserKeys[user]
	if ok {
		return keys, nil
	}
	opt := &github.ListOptions{PerPage: 25}
	for {
		githubKeys, resp, err := gc.Users.ListKeys(user, opt)
		if err != nil {
			return keys, err
		}
		for _, githubKey := range githubKeys {
			keys = append(keys, *githubKey.Key)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
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

func (gc *githubClient) getKeysOfUsersAndTeams(users, teams []string) []string {
	var keys []string
	// add members of teams to array of users
	teamMembersChan := make(chan []string)
	for _, orgName := range teams {
		go func(orgName string) {
			members, _ := gc.getMembersOfTeam(orgName)
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
			userKeys, _ := gc.getKeysOfUser(user)
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
