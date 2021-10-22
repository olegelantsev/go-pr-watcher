package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v39/github"
	"github.com/rivo/tview"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

type PullRequestMetadata struct {
	pr       *github.PullRequest
	repoName string
}

type RepoMap = map[string][]string

type Config struct {
	Repos RepoMap
	Token string
}

type SecretStorage struct {
	service string
	user    string
}

func NewSecretStorage() SecretStorage {
	return SecretStorage{
		service: "go-pr-watcher",
		user:    "github-pat",
	}
}

func (storage *SecretStorage) store(secret string) {
	keyring.Set(storage.service, storage.user, secret)
}

func (storage *SecretStorage) read() (string, error) {
	return keyring.Get(storage.service, storage.user)
}

func (storage *SecretStorage) delete() {
	err := keyring.Delete(storage.service, storage.user)
	if err != nil {
		log.Fatal(err)
	}
}

func verifyToken(client *github.Client) {
	_, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
}

func readTokenFromInput() string {
	fmt.Fprintf(os.Stdout, "GitHub Personal Token:\n")
	bytebw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytebw)
}

func loadToken() string {
	secretStorage := NewSecretStorage()
	token, err := secretStorage.read()
	if err != nil {
		if err.Error() == "secret not found in keyring" {
			token = readTokenFromInput()
			client := createGitHubClient(token)
			verifyToken(client)
			secretStorage.store(token)
		} else {
			log.Fatal(err)
		}
	}
	return token
}

func loadConfig(app *tview.Application) Config {
	var config Config

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	config.Token = loadToken()
	return config
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func listPullRequests(repos RepoMap, client *github.Client) chan PullRequestMetadata {
	channel := make(chan PullRequestMetadata)

	go func() {
		retrievePullRequests(repos, client, channel)
	}()
	return channel
}

func retrievePullRequests(repos map[string][]string, client *github.Client, channel chan PullRequestMetadata) {
	prState := &github.PullRequestListOptions{State: "open"}
	for userName, repoNames := range repos {
		for _, repoName := range repoNames {
			repo, _, err := client.Repositories.Get(context.Background(), userName, repoName)
			if err != nil {
				if strings.Contains(err.Error(), "401 Bad credentials") {
					fmt.Fprintf(os.Stderr, "Bad Credentials")
				}
				panic(err)
			}
			prs, _, err := client.PullRequests.List(context.Background(), userName, repo.GetName(), prState)
			if err != nil {
				panic(err)
			}
			for _, pr := range prs {
				channel <- PullRequestMetadata{pr, repo.GetName()}
			}
		}
	}
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func drawTable(app *tview.Application, channel chan PullRequestMetadata) *tview.Table {

	table := tview.NewTable().
		SetBorders(true)
	cols := 5
	for c := 0; c < cols; c++ {
		table.GetCell(0, c).SetSelectable(false)

		var text string
		switch c {
		case 0:
			text = "PR title"
		case 1:
			text = "state"
		case 2:
			text = "repo name"
		case 3:
			text = "user name"
		case 4:
			text = "created at"
		}
		color := tcell.ColorYellow
		table.SetCell(0, c,
			tview.NewTableCell(text).
				SetTextColor(color).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}
	prs := []PullRequestMetadata{}
	go func() {
		r := 1
		for pr := range channel {
			for c := 0; c < cols; c++ {
				color := tcell.ColorWhite
				var text string

				switch c {
				case 0:
					text = *pr.pr.Title
				case 1:
					text = *pr.pr.State
				case 2:
					text = pr.repoName
				case 3:
					text = *pr.pr.GetUser().Login
				case 4:
					text = pr.pr.CreatedAt.String()
				}

				app.QueueUpdateDraw(func() {
					table.SetCell(r, c,
						tview.NewTableCell(text).
							SetTextColor(color).
							SetAlign(tview.AlignCenter))
				})

			}
			prs = append(prs, pr)
			r += 1
		}
	}()

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, false)
		}
	}).SetSelectedFunc(func(row int, column int) {
		table.SetSelectable(false, false)
		openbrowser(*prs[row-1].pr.HTMLURL)
	})

	return table
}

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	app := tview.NewApplication()

	config := loadConfig(app)
	gitHubClient := createGitHubClient(config.Token)
	pullRequestsChannel := listPullRequests(config.Repos, gitHubClient)
	table := drawTable(app, pullRequestsChannel)
	app.SetRoot(table, true).SetFocus(table)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
