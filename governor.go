package main

import (
	"context"
	"fmt"
    "sort"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func clearConsole() {
    // fmt.Println("\033[2J")
    clearCmd := exec.Command("clear")
    clearCmd.Stdout = os.Stdout
    clearCmd.Run()
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		panic(err)
	}

    containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
    if err != nil {
        panic(err)
    }

	for {
        clearConsole()
        var containerStatusMap = make(map[string]string)
		for _, container := range containers {
			cid := container.ID
			cname := container.Names[0][1:]
			json, err := cli.ContainerInspect(ctx, cname)
			if err != nil {
				panic(err)
			}
            containerStatusMap[cname] = json.State.Status
			if json.State.Status == "exited" {
                clearConsole()
				fmt.Printf("Oh no!\n%s crashed; Restarting it now...\n", cname)
				duration, _ := time.ParseDuration("1s")
				cli.ContainerRestart(ctx, cid, &duration)
                time.Sleep(2 * time.Second)
                clearConsole()
			}
		}
        keys := make([]string, 0, len(containerStatusMap))
        for k := range containerStatusMap {
            keys = append(keys, k)
        }
        sort.Strings(keys)

        for _, k := range keys {
            fmt.Println(k, "»", containerStatusMap[k])
        }
		fmt.Printf("\nEverything's fine and dandy, keep going!")
		time.Sleep(2 * time.Second)
	}
}
