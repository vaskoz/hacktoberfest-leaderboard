package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

var (
	stderr io.Writer      = os.Stderr
	exit   func(int)      = os.Exit
	stop   chan os.Signal = make(chan os.Signal, 1)
)

type Tuple struct {
	User string
	PR   int
}

func main() {
	ctx := context.Background()

	var ghRepo, ghToken string

	//	signal.Notify(stop, os.Interrupt)

	if ghToken = os.Getenv("HFL_GH_TOKEN"); ghToken == "" {
		fmt.Fprintln(stderr, "You must set environment variable 'HFL_GH_TOKEN'")
		exit(1)
	}

	if ghRepo = os.Getenv("HFL_GH_REPO"); ghRepo == "" {
		fmt.Fprintln(stderr, "You must set environment variable 'HFL_GH_REPO'")
		exit(1)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner, repoName, readmeSha, participantsSha := setupNewRepo(ctx, client, ghRepo)
	updateStats(ctx, client, owner, repoName, readmeSha, participantsSha)

	exit(0)
}

func updateStats(ctx context.Context, client *github.Client, owner, repo, readmeSha, participantsSha string) {
	participants := map[string]int{owner: -1}
	q := "is:pr -label:invalid,spam author:%s created:2021-10-01..2021-11-01"
	msg := "Thank you for signing up for the leaderboard"
	closed := "closed"

	for { // infinite loop - need context checking here.
		// find all participants by searching for new issues.
		isr, _, _ := client.Search.Issues(ctx, fmt.Sprintf("is:issue state:open repo:%s/%s", owner, repo), &github.SearchOptions{})

		for _, iss := range isr.Issues {
			p := *iss.User.Login
			participants[p] = -1
			client.Issues.CreateComment(ctx, owner, repo, iss.GetNumber(), &github.IssueComment{
				Body: &msg,
			})
			client.Issues.Edit(ctx, owner, repo, iss.GetNumber(), &github.IssueRequest{
				State: &closed,
			})
		}

		if isr.GetTotal() != 0 {
			lst := make([]string, 0, len(participants))
			for s := range participants {
				lst = append(lst, s)
			}
			sort.StringSlice(lst).Sort()
			content := ""
			for _, s := range lst {
				content += fmt.Sprintf("%s\n", s)
			}
			opts := &github.RepositoryContentFileOptions{
				Message: github.String("participants.txt update"),
				Content: []byte(content),
				Branch:  github.String("main"),
				SHA:     github.String(participantsSha),
			}

			rcr, _, _ := client.Repositories.UpdateFile(ctx, owner, repo, "participants.txt", opts)
			participantsSha = rcr.GetContent().GetSHA()
		}

		newParticipants := make(map[string]int, len(participants))
		for user := range participants {
			isr, _, _ := client.Search.Issues(ctx, fmt.Sprintf(q, user), &github.SearchOptions{})
			// verify rules here
			validPrCount := 0
			for _, iss := range isr.Issues {
				isValid := false
				// check if PR has desired label
				for _, label := range iss.Labels {
					l := strings.ToLower(label.GetName())
					if l == "hacktoberfest-accepted" {
						isValid = true
						break
					}
				}
				// check if Repo has desired topic
				r := iss.GetHTMLURL()
				if j := strings.LastIndex(r, "/pull/"); j >= 0 {
					r = r[:j]
				}
				parts := strings.Split(r, "/")
				re := parts[len(parts)-1]
				own := parts[len(parts)-2]

				gr, _, _ := client.Repositories.Get(ctx, own, re)
				for _, topic := range gr.Topics {
					t := strings.ToLower(topic)
					if t == "hacktoberfest" {
						isValid = true
						break
					}
				}

				if isValid {
					validPrCount++
				}
			}

			newParticipants[user] = validPrCount
		}

		if !reflect.DeepEqual(participants, newParticipants) {
			sorted := Sort(newParticipants)
			// write a new file
			now := time.Now()
			fileContent := fmt.Sprintf("# Hacktoberfest 2021 Leaderboard.\n### Updated: %s\n", now.Format(time.RFC1123Z))
			for i, s := range sorted {
				fileContent += fmt.Sprintf("%d. [%s](https://github.com/%s) - %d\n", i+1, s.User, s.User, s.PR)
			}
			opts := &github.RepositoryContentFileOptions{
				Message: github.String("README.md leaderboard update"),
				Content: []byte(fileContent),
				Branch:  github.String("main"),
				SHA:     github.String(readmeSha),
			}

			rcr, _, _ := client.Repositories.UpdateFile(ctx, owner, repo, "README.md", opts)
			readmeSha = rcr.GetContent().GetSHA()
		}

		participants = newParticipants

		time.Sleep(2 * time.Minute) // 2-minutes between checks.
	}
}

func Sort(p map[string]int) []Tuple {
	res := make([]Tuple, 0, len(p))

	for k, v := range p {
		res = append(res, Tuple{k, v})
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].PR > res[j].PR {
			return true
		} else if res[i].PR == res[j].PR && res[i].User < res[j].User {
			return true
		}
		return false
	})

	return res
}

func setupNewRepo(ctx context.Context, client *github.Client, ghRepo string) (string, string, string, string) {
	F := false
	T := true
	r := &github.Repository{Name: &ghRepo, Private: &F, HasIssues: &T, HasWiki: &F,
		HasPages: &F, HasProjects: &F, HasDownloads: &F}
	repo, _, err := client.Repositories.Create(ctx, "", r)
	if err != nil {
		fmt.Fprintln(stderr, repo, err)
	}

	time.Sleep(5 * time.Second)

	now := time.Now()
	fileContent := fmt.Sprintf("# Hacktoberfest 2021 Leaderboard. Updated: %s\n", now.Format(time.RFC1123Z))
	opts := &github.RepositoryContentFileOptions{
		Message: github.String("Setup README.md"),
		Content: []byte(fileContent),
		Branch:  github.String("main"),
	}

	rcr, _, _ := client.Repositories.CreateFile(ctx, *repo.Owner.Login, *repo.Name, "README.md", opts)

	opts1 := &github.RepositoryContentFileOptions{
		Message: github.String("Setup blank participants.txt"),
		Content: []byte(*repo.Owner.Login),
		Branch:  github.String("main"),
	}

	rcr2, _, _ := client.Repositories.CreateFile(ctx, *repo.Owner.Login, *repo.Name, "participants.txt", opts1)

	return *repo.Owner.Login, *repo.Name, *rcr.GetContent().SHA, *rcr2.GetContent().SHA
}
