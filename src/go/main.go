package main

import (
	"github.com/isoron/notes/notes"
	"github.com/jcelliott/lumber"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"time"
)

var version string
var dataDir string

func main() {
	app := cli.NewApp()
	app.Name = "notes"
	app.Usage = "a simple wiki"
	app.Version = version
	app.Compiled = time.Now()
	app.Action = func(c *cli.Context) error {
		dataDir = c.GlobalString("data")
		site := notes.Site{
			DataDir:                dataDir,
			AllowFileUploads:       c.GlobalBool("allow-file-uploads"),
			MaxUploadSizeInMB:      c.GlobalUint("max-upload-mb"),
			Logger:                 lumber.NewConsoleLogger(lumber.TRACE),
			MaxPageContentSizeInMB: c.GlobalUint("max-document-length"),
		}
		err := site.Migrate()
		if err != nil {
			log.Fatal(err)
		}
		site.Run(
			c.GlobalString("host"),
			c.GlobalString("port"),
		)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "data",
			Value: "data",
			Usage: "data folder to use",
		},
		cli.StringFlag{
			Name:  "host",
			Value: "0.0.0.0",
			Usage: "host to use",
		},
		cli.StringFlag{
			Name:  "port,p",
			Value: "8050",
			Usage: "port to use",
		},
		cli.BoolFlag{
			Name:  "allow-file-uploads",
			Usage: "Enable file uploads",
		},
		cli.UintFlag{
			Name:  "max-upload-mb",
			Value: 2,
			Usage: "Largest file upload (in mb) allowed",
		},
		cli.UintFlag{
			Name:  "max-document-length",
			Value: 100000000,
			Usage: "Largest wiki page (in characters) allowed",
		},
	}
	app.Run(os.Args)
}
