package main

import (
	"code/internal/crawler"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:      "hexlet-go-crawler",
		Usage:     "analyze a website structure",
		ArgsUsage: "<url>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "url",
			},
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "depth",
				Value: 10,
				Usage: "crawl depth",
				//Category: "GLOBAL OPTIONS:",
			},
			&cli.IntFlag{
				Name:  "retries",
				Value: 1,
				Usage: "number of retries for failed requests",
				//Category: "GLOBAL OPTIONS:",,
			},
			&cli.DurationFlag{
				Name:     "delay",
				Value:    0,
				Usage:    "delay between requests (example: 200ms, 1s)",
				Category: "GLOBAL OPTIONS:",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			options := crawler.Options{
				Client:  &http.Client{},
				Depths:  c.Int("depth"),
				Retries: c.Int("retries"),
				Delay:   c.Duration("delay"),
				Url:     c.StringArg("url"),
			}

			report, err := crawler.Analyze(ctx, options)
			if err != nil {
				return err
			}
			fmt.Println(report)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
