package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/thomaslefeuvre/bandcamp-tools/bandcamp"
	"github.com/thomaslefeuvre/bandcamp-tools/dig"
	"github.com/thomaslefeuvre/bandcamp-tools/gmail"
	"github.com/urfave/cli/v2"
)

const (
	bandcampUsername = "thomaslefeuvre"
	outDir           = "data"
	secretDir        = "secrets"
)

func main() {
	ctx := context.Background()

	app := &cli.App{
		Name:  "dig",
		Usage: "",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:        "n",
				DefaultText: "10",
				Usage:       "Number of items to collect",
			},
			&cli.BoolFlag{
				Name:  "save",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "open",
				Value: false,
			},
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"in"},
			},
			&cli.StringSliceFlag{
				Name:    "strategy",
				Aliases: []string{"s"},
			},
		},
		Action: func(c *cli.Context) error {
			n := c.Int64("n")
			if n > int64(500) {
				return fmt.Errorf("n cannot exceed 500")
			}

			save := c.Bool("save")
			open := c.Bool("open")

			inputFile := c.String("input")

			var result *dig.Result
			var err error
			if inputFile != "" {
				result, err = dig.NewResultFromFile(inputFile)
				if err != nil {
					return err
				}
			} else {
				strategies := c.StringSlice("strategy")
				var collectors []dig.Collector
				for _, s := range strategies {
					if s == "gmail" {
						svc, err := gmail.NewService(ctx, secretDir)
						if err != nil {
							return err
						}
						collectors = append(collectors, dig.NewGmail(svc))
					}
					if s == "bandcamp" {
						me := &bandcamp.User{Name: bandcampUsername}
						collectors = append(collectors, dig.NewWishlist(me))
					}
				}
				fmt.Println(n)
				dig := dig.New(n, collectors...)
				result, err = dig.Run()
				if err != nil {
					return err
				}
			}

			if save {
				_, err := result.Save(outDir)
				if err != nil {
					return err
				}
			}

			if open {
				err := result.OpenInBrowser()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
