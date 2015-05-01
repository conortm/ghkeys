# ghkeys [![Build Status](https://travis-ci.org/conortm/ghkeys.svg)](https://travis-ci.org/conortm/ghkeys) [![Coverage Status](https://coveralls.io/repos/conortm/ghkeys/badge.svg?branch=master)](https://coveralls.io/r/conortm/ghkeys?branch=master)

[ghkeys](https://github.com/conortm/ghkeys) is a simple command line tool for syncing server users' authorized SSH keys with those of one or more GitHub accounts.

Via configuration, specify individual GitHub users and/or entire teams, whose SSH keys should either be output directly or written to server users' `authorized_keys` files.

## Install

```sh
$ go get github.com/conortm/ghkeys
```

## Configure

Create a `config.yml` file like:

```yaml
---
# Replace with your own GitHub token:
github_token: my_github_token
# Array of server usernames and the GitHub source(s) of their authorized keys:
users:
  - username: superadmin
    github_users:
      - github_user_1
    github_teams:
      # Specify teams by Org name and team name, separated by a /
      - MyOrg/Team 1
      - MyOrg/Team 2
  - username: admin
    github_users:
      - github_user_1
      - github_user_2
    github_teams:
      - MyOtherOrg/Team 3
```

## Execute

To print all keys:

```sh
$ ghkeys -config="/path/to/config.yml"
```

Pass single username argument to print only that user's keys, for example, when using `AuthorizedKeysCommand`:

```sh
$ ghkeys -config="/path/to/config.yml" superadmin
```

Use the `-write` flag to write keys to users' `authorized_keys` files:

```sh
$ ghkeys -config="/path/to/config.yml" -write
```

## TODO

* Implement https://github.com/sourcegraph/apiproxy
* Validate config file
* Better error handling
* Tests
