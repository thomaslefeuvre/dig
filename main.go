package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/thomaslefeuvre/digg/bandcamp"
	"github.com/thomaslefeuvre/digg/dig"
	"github.com/thomaslefeuvre/digg/gmail"
	"github.com/urfave/cli/v2"
)

const (
	bandcampUsername = "thomaslefeuvre"
	outDir           = "data"
	secretDir        = "secrets"
	maxToOpen        = 50
)

func main() {
	ctx := context.Background()

	app := &cli.App{
		Name:  "digg",
		Usage: "A tool to streamline your digging process.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"dry", "d"},
				Usage:   "Do not commit changes to collection.",
			},
		},
		Action: func(c *cli.Context) error {
			dryRun := c.Bool("dry-run")

			var collection *dig.Collection
			var err error

			collection, err = dig.LoadCollection(outDir)
			if err != nil {
				log.Printf("No collection exists, creating a new one: %v", err)
				collection = dig.NewCollection(outDir)
			}

			gc, err := gmail.NewService(ctx, secretDir)
			if err != nil {
				log.Printf("Unable to create gmail service: %v", err)
				return err
			}

			bc := &bandcamp.User{Name: bandcampUsername}

			collectors := []dig.Collector{dig.NewGmail(gc), dig.NewWishlist(bc)}

			dig := dig.New(collectors...)
			collection = dig.UpdateCollection(collection)

			if !dryRun {
				names, err := collection.Save()
				if err != nil {
					log.Printf("Unable to save collection: %v", err)
				}

				for _, name := range names {
					log.Printf("Saved collection to: %v", name)
				}
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "open",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "query",
						Aliases:     []string{"q"},
						Usage:       "Query to filter collection",
						DefaultText: "None",
					},
					&cli.IntFlag{
						Name:        "n",
						DefaultText: "10",
						Usage:       "Number of items to collect",
					},
				},
				Action: func(c *cli.Context) error {
					n := c.Int("n")
					if n == 0 {
						n = maxToOpen
					}
					if n > maxToOpen {
						log.Printf("n exceeds maximum of %d, setting n=%d\n", maxToOpen, maxToOpen)
						n = maxToOpen
					}

					q := c.String("query")

					collection, err := dig.LoadCollection(outDir)
					if err != nil {
						log.Fatalln("Unable to load collection")
					}

					if q == "" {
						err = collection.Open(n)
					} else {
						err = collection.OpenFilter(q, n)
					}
					if err != nil {
						log.Printf("Unable to open collection in browser: %v", err)
					}

					if !c.Bool("dry-run") {
						_, err := collection.Save()
						if err != nil {
							log.Printf("Unable to save collection: %v", err)
						}
					}

					return nil
				},
			},
			{
				Name: "list",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "query",
						Aliases:     []string{"q"},
						Usage:       "Query to filter collection",
						DefaultText: "None",
					},
				},
				Action: func(c *cli.Context) error {
					query := c.String("q")

					collection, err := dig.LoadCollection(outDir)
					if err != nil {
						log.Fatalln("Unable to load collection")
					}

					if collection.Size() == 0 {
						fmt.Println("Collection is empty")
						return nil
					}

					var items []string
					if query == "" {
						items = collection.All()
					} else {
						items = collection.Filter(query)
					}

					for _, item := range items {
						fmt.Println(item)
					}

					return nil
				},
			},
			{
				Name: "info",
				Action: func(c *cli.Context) error {
					collection, err := dig.LoadCollection(outDir)
					if err != nil {
						log.Fatalln("Unable to load collection")
					}

					if collection.Size() == 0 {
						fmt.Println("Collection is empty")
						return nil
					}

					fmt.Println("Collection size:", collection.Size())

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
