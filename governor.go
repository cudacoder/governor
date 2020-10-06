package main

import (
	// "fmt"
	"bufio"
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func RestartDockerContainer(cli *client.Client, ctx context.Context, cid string) {
	duration, _ := time.ParseDuration("1s")
	cli.ContainerRestart(ctx, cid, &duration)
}

func ContainerStatusArray(cli *client.Client, ctx context.Context) []string {
	sliceOfContainers := []string{}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	containerStatusMap := make(map[string]string)

	for _, container := range containers {
		cname := container.Names[0][1:]
		json, err := cli.ContainerInspect(ctx, cname)
		if err != nil {
			panic(err)
		}
		containerStatusMap[cname] = json.State.Status
	}
	keys := make([]string, 0, len(containerStatusMap))
	for k := range containerStatusMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sliceOfContainers = append(sliceOfContainers, k+" » "+containerStatusMap[k])
	}
	return sliceOfContainers
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))

	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	defer ui.Close()

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	l := widgets.NewList()
	l.Title = "Running Services"
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, 60, 20)
	l.Rows = ContainerStatusArray(cli, ctx)

	p2 := widgets.NewParagraph()
	p2.Text = ""
	p2.Border = false
	p2.SetRect(50, 10, 75, 10)
	p2.TextStyle.Fg = ui.ColorClear

	grid.Set(
		ui.NewRow(1.0/2, ui.NewCol(1.0, l)),
		ui.NewRow(1.0/2, ui.NewCol(1.0, p2)),
	)

	tickerCount := 1
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "r":
				s := strings.Fields(l.Rows[l.SelectedRow])[0]
				RestartDockerContainer(cli, ctx, s)
			case "j":
				l.ScrollDown()
			case "k":
				l.ScrollUp()
			case "<C-d>":
				l.ScrollHalfPageDown()
			case "<C-u>":
				l.ScrollHalfPageUp()
			case "<C-f>":
				l.ScrollPageDown()
			case "<C-b>":
				l.ScrollPageUp()
			case "G":
				l.ScrollBottom()
			}
			ui.Render(grid)
		case <-ticker:
			clogs := []string{}
			l.Rows = ContainerStatusArray(cli, ctx)
			selected := strings.Fields(l.Rows[l.SelectedRow])[0]
			reader, err := cli.ContainerLogs(ctx, selected, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close()

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				clogs = append(clogs, scanner.Text())
			}
			for i := len(clogs)/2 - 1; i >= 0; i-- {
				opp := len(clogs) - 1 - i
				clogs[i], clogs[opp] = clogs[opp], clogs[i]
			}
			if len(clogs) != 0 {
				p2.Text = strings.Join(clogs[:5], "\n")
			}

			tickerCount += 1
			ui.Render(grid)
		}
	}
}
