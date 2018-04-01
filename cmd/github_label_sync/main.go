// Copyright 2017 github-label-sync Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	fAuthtoken   string
	fGithubOwner string
	fGithubRepo  string
	labelColors  = map[string]string{
		"review/triage": "924cb2",

		"P0": "b60205",
		"P1": "d93f0b",
		"P2": "e99695",
		"P3": "c2e0c6",
		"P4": "c5def5",

		"1":  "fef2c0",
		"2":  "f9d0c4",
		"4":  "e99695",
		"8":  "d93f0b",
		"16": "b60205",

		"backlog": "f7d74a",
		"Task":    "1d76db",
		"Story":   "1c8300",
		"Epic":    "3E4B9E",
		"S":       "0c508c",
	}
)

const (
	usage = `
Usage of %s:

Github receiver requires a github --authtoken and target github --owner and
--repo names.

`
)

func init() {
	flag.StringVar(&fAuthtoken, "authtoken", "", "Oauth2 token for access to github API.")
	flag.StringVar(&fGithubOwner, "owner", "", "The github user or organization name.")
	flag.StringVar(&fGithubRepo, "repo", "", "The repository where issues are created.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
}

// A Client manages communication with the Github API.
type Client struct {
	// githubClient is an authenticated client for accessing the github API.
	GithubClient *github.Client
	// owner is the github project (e.g. github.com/<owner>/<repo>).
	owner string
	// repo is the github repository under the above owner.
	repo string
}

// NewClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewClient(owner, repo, authToken string) *Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	client := &Client{
		GithubClient: github.NewClient(oauth2.NewClient(ctx, tokenSource)),
		owner:        owner,
		repo:         repo,
	}
	return client
}

func pString(s string) *string {
	return &s
}

func newLabel(name, color string) *github.Label {
	return &github.Label{
		Name:  &name,
		Color: &color,
	}
}

type localClient Client

func (l localClient) syncLabel(current map[string]string, name, color string) error {
	colorBefore, found := current[name]

	switch {
	case !found:
		fmt.Printf("Creating: label with color %s %s\n", color, name)
		_, _, err := l.GithubClient.Issues.CreateLabel(
			context.Background(), fGithubOwner, fGithubRepo, newLabel(name, color))
		if err != nil {
			fmt.Println("Create failed", err)
			return err
		}
		// fmt.Println(resp)
		// fmt.Println(labelAfter)
	case found && color == colorBefore:
		fmt.Printf("Verified: color up to date for %s\n", name)
	case found && color != colorBefore:
		fmt.Printf("Updating: color %s to %s for %s\n", colorBefore, color, name)
		_, _, err := l.GithubClient.Issues.EditLabel(
			context.Background(), fGithubOwner, fGithubRepo, name, newLabel(name, color))
		if err != nil {
			fmt.Println("Update failed", err)
			return err
		}
		// fmt.Println(resp)
		// fmt.Println(labelAfter)
	}
	return nil
}

func (l localClient) deleteLabel(name string) error {
	fmt.Printf("Deleting: label %s\n", name)
	_, err := l.GithubClient.Issues.DeleteLabel(
		context.Background(), fGithubOwner, fGithubRepo, name)
	if err != nil {
		fmt.Println("")
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if fAuthtoken == "" || fGithubOwner == "" || fGithubRepo == "" {
		flag.Usage()
		os.Exit(1)
	}
	client := (*localClient)(NewClient(fGithubOwner, fGithubRepo, fAuthtoken))
	foundLabels, resp, err := client.GithubClient.Issues.ListLabels(
		context.Background(), fGithubOwner, fGithubRepo, nil)
	if err != nil {
		log.Fatal(err)
	}

	knownLabels := make(map[string]string, len(foundLabels))
	fmt.Println(resp)
	for _, l := range foundLabels {
		knownLabels[l.GetName()] = l.GetColor()
	}

	for name, color := range labelColors {
		err := client.syncLabel(knownLabels, name, color)
		if err != nil {
			log.Fatal(err)
			break
		}
	}
	// TODO: delete extra known labels missing from LabelColors.
	for name := range knownLabels {
		_, found := labelColors[name]
		if !found {
			fmt.Printf("Ignoring: found unknown label %s\n", name)
			// err := client.deleteLabel(name)
			// if err != nil {
			// log.Fatal(err)
			// break
			// }
		}
	}
}
