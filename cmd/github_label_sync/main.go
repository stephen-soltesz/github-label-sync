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
	"github.com/stephen-soltesz/github-label-sync/issues"
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

func pString(s string) *string {
	return &s
}

func newLabel(name, color string) *github.Label {
	return &github.Label{
		Name:  &name,
		Color: &color,
	}
}

func main() {
	flag.Parse()
	if fAuthtoken == "" || fGithubOwner == "" || fGithubRepo == "" {
		flag.Usage()
		os.Exit(1)
	}
	client := issues.NewClient(fGithubOwner, fGithubRepo, fAuthtoken)
	foundLabels, resp, err := client.GithubClient.Issues.ListLabels(context.Background(), fGithubOwner, fGithubRepo, nil)
	if err != nil {
		log.Fatal(err)
	}
	knownLabels := map[string]string{}
	fmt.Println(resp)
	for i, l := range foundLabels {
		knownLabels[l.GetName()] = l.GetColor()
		fmt.Println(i, l.GetColor(), l.GetName())
	}

	for name, color := range labelColors {
		if _, ok := knownLabels[name]; ok {
			fmt.Println("Found:", name)
			// TODO: update colors if needed.
			continue
		}

		// Create label.
		label := newLabel(name, color)
		fmt.Println("Creating:", label)

		l, resp, err := client.GithubClient.Issues.CreateLabel(context.Background(), fGithubOwner, fGithubRepo, label)
		if err != nil {
			log.Fatal("Fatal:" + err.Error())
		}
		fmt.Println(l)
		fmt.Println(resp)
	}
	// TODO: delete extra known labels missing from LabelColors.
}
