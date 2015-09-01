package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const usageMessage = `
'ghkeys' uses the GitHub API to get the SSH keys of individual users and/or
members of teams and either print them or write them to authorized_keys files.

Pass a single 'username' argument to only print/write keys for that user.

Usage:
  ghkeys [-config="/path/to/config.yml"] [-write] [username]
`

var (
	// VERSION swapped out by `go build -ldflags "-X main.VERSION 1.2.3"`
	VERSION        = "0.0.5"
	configFilename = flag.String("config", "config.yml", "Path to yaml config file")
	debug          = flag.Bool("d", false, "Add debugging output")
	showVersion    = flag.Bool("version", false, "Display the version number")
	writeToFile    = flag.Bool("write", false, "Write keys to users' authorized_keys files")
)

func printVersion() {
	fmt.Printf("ghkeys version %s\n", VERSION)
}

func usage() {
	printVersion()
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

func getAuthorizedKeysFilename(userHomeDir string) string {
	authorizedKeysFilename := filepath.Join(userHomeDir, ".ssh", "authorized_keys")
	return authorizedKeysFilename
}

func getKeysOutput(keys []string) string {
	keysOutput := strings.Join(keys, "\n")
	return keysOutput
}

func writeKeysToUserAuthorizedKeysFile(keys []string, userHomeDir string) error {
	authorizedKeysFilename := getAuthorizedKeysFilename(userHomeDir)
	authorizedKeysFile, err := os.Create(authorizedKeysFilename)
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

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	config, err := newConfig(*configFilename)
	check(err)
	// TODO: validate config values

	client := newGithubClient(config.GithubToken)

	if *debug {
		if rate, _, err := client.RateLimit(); err == nil {
			log.Printf("GitHub API Rate Limit: %#v\n", rate)
		}
	}

	usernamesKeys := getUsernamesKeys(config, client, flag.Arg(0))

	for username, keys := range usernamesKeys {
		if *writeToFile {
			if user, err := user.Lookup(username); err == nil {
				err = writeKeysToUserAuthorizedKeysFile(keys, user.HomeDir)
				check(err)
			} else {
				log.Printf("User '%s' not found, no keys written.\n", username)
			}
		} else {
			fmt.Println(getKeysOutput(keys))
		}
	}
}
