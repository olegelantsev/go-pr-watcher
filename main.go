package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/google/go-github/v39/github"
	"github.com/rivo/tview"
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

func loadConfig() *Config {
	var config Config
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return &config
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
		repos, _, err := client.Repositories.List(context.Background(), userName, nil)
		if err != nil {
			panic(err)
		}
		log.Println(repoNames)
		for _, repo := range repos {
			if contains(repoNames, repo.GetName()) {
				prs, _, err := client.PullRequests.List(context.Background(), userName, repo.GetName(), prState)
				if err != nil {
					panic(err)
				}
				for _, pr := range prs {
					pullRequests = append(pullRequests, PullRequestMetadata{pr, repo.GetName()})
				}
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

func main() {
	config := loadConfig()
	prs := listPullRequests(config)

	app := tview.NewApplication()
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
	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
