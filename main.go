package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const usageMessage = "" + `
'ghkeys' uses the GitHub API to get the SSH keys of individual users and/or
members of teams and either print them or write them to authorized_keys files.

Pass a single 'username' argument to only print/write keys for that user.
`

var (
	configFilename = flag.String("config", "config.yml", "Path to yaml config file")
	writeToFile    = flag.Bool("write", false, "Write keys to users' authorized_keys files")
)

func usage() {
	fmt.Println(usageMessage)
	fmt.Println("Flags:")
	flag.PrintDefaults()
	os.Exit(2)
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

func main() {
	flag.Usage = usage
	flag.Parse()

	config, err := newConfig(*configFilename)
	check(err)

	singleUsername := flag.Arg(0)

	// TODO: validate config, including that usernames exist on server.

	gc := newGithubClient(config.GithubToken)

	usernameCount := 0
	usernameKeysChan := make(chan usernameKeys)
	for _, user := range config.Users {
		if singleUsername == "" || singleUsername == user.Username {
			usernameCount++
			go func(username string, users, teams []string) {
				keys := gc.getKeys(users, teams)
				usernameKeysChan <- usernameKeys{username: username, keys: keys}
			}(user.Username, user.GithubUsers, user.GithubTeams)
		}
	}
	for i := 0; i < usernameCount; i++ {
		usernameKey := <-usernameKeysChan
		authorizedKeysOutput := strings.Join(usernameKey.keys, "\n")
		if *writeToFile {
			authorizedKeysFilename := fmt.Sprintf("/home/%s/.ssh/authorized_keys", usernameKey.username)
			f, err := os.Create(authorizedKeysFilename)
			check(err)
			defer f.Close()
			_, err = f.WriteString(authorizedKeysOutput)
			check(err)
		} else {
			fmt.Println(authorizedKeysOutput)
		}
	}
}
