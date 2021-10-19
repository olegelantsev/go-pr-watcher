package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"

	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v39/github"
	"github.com/rivo/tview"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type PullRequestMetadata struct {
	pr       *github.PullRequest
	repoName string
}

type Config struct {
	Repos map[string][]string
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

func loadConfig(app *tview.Application, onSuccess func(config *Config)) {
	var config Config

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	wm := winman.NewWindowManager()
	secretStorage := NewSecretStorage()
	token, err := secretStorage.read()
	if err != nil {
		if err.Error() == "secret not found in keyring" {

			passwordField := tview.NewInputField().
				SetLabel("Token").
				SetText("").
				SetFieldWidth(30).
				SetMaskCharacter('*').
				SetChangedFunc(func(text string) {
					token = text
				}).SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyCR {
					secretStorage.store(token)
					config.Token = token
					onSuccess(&config)
				} else {
					app.Stop()
				}
			})

			form := tview.NewForm().AddFormItem(passwordField)
			window := wm.NewWindow().
				Show().
				SetRoot(form).
				SetDraggable(true).
				SetResizable(true).
				SetTitle("GitHub Personal token").
				AddButton(&winman.Button{
					Symbol:  'X',
					OnClick: func() { app.Stop() },
				})
			window.SetRect(5, 5, 50, 5)

			app.SetRoot(wm, true).EnableMouse(true)
		} else {
			log.Fatal(err)
		}
	} else {
		config.Token = token
		onSuccess(&config)
	}

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func listPullRequests(config *Config) []PullRequestMetadata {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	pullRequests := []PullRequestMetadata{}
	prState := &github.PullRequestListOptions{State: "open"}
	for userName, repoNames := range config.Repos {
		for _, repoName := range repoNames {
			repo, _, err := client.Repositories.Get(context.Background(), userName, repoName)
			if err != nil {
				panic(err)
			}
			prs, _, err := client.PullRequests.List(context.Background(), userName, repo.GetName(), prState)
			if err != nil {
				panic(err)
			}
			for _, pr := range prs {
				pullRequests = append(pullRequests, PullRequestMetadata{pr, repo.GetName()})
			}
		}
	}

	return pullRequests
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

func drawTable(config *Config, app *tview.Application) *tview.Table {
	prs := listPullRequests(config)
	table := tview.NewTable().
		SetBorders(true)
	cols, rows := 5, len(prs)+1
	for c := 0; c < cols; c++ {
		table.GetCell(0, c).SetSelectable(false)
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			color := tcell.ColorWhite
			var text string
			if c < 1 || r < 1 {
				color = tcell.ColorYellow
			}
			if r == 0 {
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
			} else {
				pr := prs[r-1]
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

			}

			table.SetCell(r, c,
				tview.NewTableCell(text).
					SetTextColor(color).
					SetAlign(tview.AlignCenter))
		}
	}
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
	app := tview.NewApplication()

	loadConfig(app, func(config *Config) {
		table := drawTable(config, app)
		app.SetRoot(table, true)
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
