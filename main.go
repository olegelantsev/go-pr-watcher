package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
		if row > 0 {
			table.SetSelectable(false, false)
			openbrowser(*prs[row-1].pr.HTMLURL)
		}
	})

	return table
}

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	app := tview.NewApplication()

	config := loadConfig(app)
	gitHubWrapper := NewGitHubWrapper(config.Token)
	pullRequestsChannel := gitHubWrapper.ListPullRequests(config.Repos)
	table := drawTable(app, pullRequestsChannel)
	app.SetRoot(table, true).SetFocus(table)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
