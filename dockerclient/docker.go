package dockerclient

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func RestartConatiners(app string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("error while getting client %v", err)
		return err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		fmt.Printf("error while getting containers %v", err)
		return err
	}

	// Loop through the containers and find the one with the matching name
	for _, container := range containers {
		for _, name := range container.Names {
			// Docker names have a leading slash, so we remove it
			match, err := regexp.MatchString(app, name[1:])
			if err != nil {
				fmt.Printf("error while matching %v", err)
				continue
			}
			fmt.Printf("matched:  %v\n", match)
			if match {
				// Restart the container
				err = cli.ContainerRestart(ctx, container.ID, nil)
				if err != nil {
					fmt.Printf("error while restarting container %v", err)
					return err
				}
				log.Println("Container has been restarted")
			}
		}
	}

	return nil
}
