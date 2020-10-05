package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		panic(err)
	}

	for {
		clearCmd := exec.Command("clear")
		clearCmd.Stdout = os.Stdout
		clearCmd.Run()
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}
		for _, container := range containers {
			cid := container.ID
			cname := container.Names[0][1:]
			json, err := cli.ContainerInspect(ctx, cid)
			if err != nil {
				panic(err)
			}
			if json.State.ExitCode != 0 {
				fmt.Printf("Oh no!\n%s crashed; Restarting it now...", cname)
				duration, _ := time.ParseDuration("1s")
				cli.ContainerRestart(ctx, cid, &duration)
			}
			fmt.Printf("%s » %s (%s)\n", container.Names[0][1:], container.State, container.Image)
		}
		fmt.Printf("\nEverything's fine and dandy, keep going!")
		time.Sleep(4 * time.Second)
	}
}
