package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const usageMessage = `
'ghkeys' uses the GitHub API to get the SSH keys of individual users and/or
members of teams and either print them or write them to authorized_keys files.

Pass a single 'username' argument to only print/write keys for that user.
`

var (
	configFilename = flag.String("config", "config.yml", "Path to yaml config file")
	debug          = flag.Bool("d", false, "Add debugging output")
	writeToFile    = flag.Bool("write", false, "Write keys to users' authorized_keys files")
)

func usage() {
	fmt.Println(usageMessage + "\nFlags:")
	flag.PrintDefaults()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type usernameKeys struct {
	username string
	keys     []string
}

func getUsernamesKeys(config config, client *githubClient, singleUsername string) map[string][]string {
	usernamesKeys := make(map[string][]string)
	usernameCount := 0
	usernameKeysChan := make(chan usernameKeys)
	for _, user := range config.Users {
		if singleUsername == "" || singleUsername == user.Username {
			usernameCount++
			go func(username string, users, teams []string) {
				keys := client.getKeysOfUsersAndTeams(users, teams)
				usernameKeysChan <- usernameKeys{username: username, keys: keys}
			}(user.Username, user.GithubUsers, user.GithubTeams)
		}
	}
	for i := 0; i < usernameCount; i++ {
		tempUsernameKeys := <-usernameKeysChan
		usernamesKeys[tempUsernameKeys.username] = tempUsernameKeys.keys
	}
	return usernamesKeys
}

func getKeysOutput(keys []string) string {
	return strings.Join(keys, "\n")
}

func writeKeysToFile(keys []string, filename string) error {
	authorizedKeysFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer authorizedKeysFile.Close()
	keysOutput := getKeysOutput(keys) + "\n"
	_, err = authorizedKeysFile.WriteString(keysOutput)
	return err
}

func main() {
	flag.Usage = usage
	flag.Parse()

	config, err := newConfig(*configFilename)
	check(err)
	// TODO: validate config values

	client := newGithubClient(config.GithubToken)

	if *debug {
		if rate, _, err := client.RateLimit(); err == nil {
			fmt.Printf("GitHub API Rate Limit: %#v\n", rate)
		}
	}

	usernamesKeys := getUsernamesKeys(config, client, flag.Arg(0))

	for username, keys := range usernamesKeys {
		if *writeToFile {
			// TODO: Is this always the path?
			err := writeKeysToFile(keys, fmt.Sprintf("/home/%s/.ssh/authorized_keys", username))
			check(err)
		} else {
			fmt.Println(getKeysOutput(keys))
		}
	}
}
