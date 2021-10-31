# hacktoberfest-leaderboard
a leaderboard system that can run on any Internet enabled device.

# How it works
* The app updates the leaderboard on README.md in your Github repo.
* The update happens every *2 minutes*.
* **Joining Leaderboard:** Simply have that participant create a new issue on your leaderboard repo with their Github account.

# Step-by-step Instructions
1. Login to your account at [Github.](https://github.com/) If you don't already have an account, [create a free one.](https://github.com/join)
1. Create a [Personal Access Token.](https://github.com/settings/tokens)
1. Copy down the personal access token for future use. Now head over to the Usage section.
# Usage

You need environment variables to operate this program successfully.

```
HFL_GH_TOKEN=yourGithubPersonalToken HFL_GH_REPO=nameOfLeaderboardRepoToCreate ./hacktoberfest-leaderboard
```

## Restart existing leaderboard

Now you can re-run the same command and it will pick up where it left off on an existing repository.

# Setup multiple leaderboards
You're able to run as many leaderboards as you want, however, you'll be limited by Github's API rate limits. As of this writing, Github has a limit of 5,000 reqs/hour.

To run a different leaderboard at the same time, simply start another instance with the `HFL_GH_REPO` variable set to a non-existent Github repo name. You can setup as many unique leaderboard repos as you'd like.
