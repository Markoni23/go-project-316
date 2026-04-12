package main

import (
	"code/internal/crawler"
	"code/internal/crawler/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
				Value: 3,
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
			&cli.DurationFlag{
				Name:     "timeout",
				Value:    15 * time.Second,
				Usage:    "per-request timeout (default: 15s)",
				Category: "GLOBAL OPTIONS:",
			},
			&cli.StringFlag{
				Name:     "user-agent",
				Value:    "",
				Usage:    "custom user agent",
				Category: "GLOBAL OPTIONS:",
			},
			&cli.IntFlag{
				Name:     "workers",
				Value:    4,
				Usage:    "number of concurrent workers",
				Category: "GLOBAL OPTIONS:",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			options := models.Options{
				Client:       &http.Client{},
				Depth:        c.Int("depth"),
				Retries:      c.Int("retries"),
				Delay:        c.Duration("delay"),
				Url:          c.StringArg("url"),
				Timeout:      c.Duration("timeout"),
				UserAgent:    c.String("user-agent"),
				WorkersCount: c.Int("workers"),
			}
			options.Url = "https://hexlet.ru"
			//res, err := options.Client.Get("https://islod.obrnadzor.gov.ru/rlic/details/5df0362b-ec43-d1e4-cfea-1aa7ff69cad9/")
			report, err := crawler.Analyze(ctx, options)
			if err != nil {
				return err
			}
			//fmt.Println(res.StatusCode)

			fmt.Println(report)
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
