package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

var (
	err  io.Writer      = os.Stderr
	exit func(int)      = os.Exit
	stop chan os.Signal = make(chan os.Signal, 1)
)

func main() {
	ctx := context.Background()

	var ghRepo, ghToken string

	signal.Notify(stop, os.Interrupt)

	if ghToken = os.Getenv("HFL_GH_TOKEN"); ghToken == "" {
		fmt.Fprintln(err, "You must set environment variable 'HFL_GH_TOKEN'")
		exit(1)
	}

	if ghRepo = os.Getenv("HFL_GH_REPO"); ghRepo == "" {
		fmt.Fprintln(err, "You must set environment variable 'HFL_GH_REPO'")
		exit(1)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	_ = client

	exit(0)
}
