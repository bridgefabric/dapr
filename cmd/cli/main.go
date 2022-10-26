package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/dapr/dapr/pkg/config"
	"github.com/dapr/dapr/pkg/p2p"
)

// BuildDate: Binary file compilation time
// BuildVersion: Binary compiled GIT version
var (
	BuildDate    string
	BuildVersion string
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate)
	app := cli.NewApp()
	app.Name = "bridge cli"
	app.Description = "bridge cli"

	//app.Flags = []cli.Flag{
	//	cli.StringFlag{
	//		Name:     "config,c",
	//		Usage:    "配置文件",
	//		Required: false,
	//		Value:    "config/config.yaml",
	//	},
	//}
	app.Commands = []cli.Command{
		{
			Name:    "stack",
			Aliases: []string{"s"},
			Usage:   "stack this node to get the right of running wasm",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Required: true,
					Name:     "node",
					Usage:    "node id",
				},
			},
			Action: func(cCtx *cli.Context) error {

				return nil
			},
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "upload a wasm file to ipfs",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Required: true,
					Name:     "file",
					Usage:    "wasm binary file",
				},
			},
			Action: func(cCtx *cli.Context) error {

				return nil
			},
		},
		{
			Name:  "config",
			Usage: "config",
			Subcommands: []cli.Command{
				{
					Name:  "new",
					Usage: "create a new config",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "o",
							Value: "config.yaml",
							Usage: "default config file",
						},
					},
					Action: func(c *cli.Context) error {
						fileName := c.String("o")
						prvkey, nodeid, err := p2p.CreatePrivateKey()
						checkErr(err)
						conf := config.Interface{
							ID:         nodeid,
							PrivateKey: string(prvkey),
						}
						out, err := yaml.Marshal(&conf)
						checkErr(err)
						err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
						checkErr(err)

						f, err := os.Create(fileName)
						checkErr(err)

						// Write out config to file.
						_, err = f.Write(out)
						checkErr(err)

						err = f.Close()
						checkErr(err)
						fmt.Printf("Initialized new config at %s\n", fileName)
						return nil
					},
				},
				{
					Name:  "check",
					Usage: "check a config",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "c",
							Value: "config.yaml",
							Usage: "default config file",
						},
					},
					Action: func(c *cli.Context) error {
						fileName := c.String("c")
						conf, err := readConfig(fileName)
						if err != nil {
							return err
						}
						nodeid, err := p2p.GetNodeIDFromPrivateKey([]byte(conf.PrivateKey))
						checkErr(err)
						if nodeid != conf.ID {
							return fmt.Errorf("node id not matchwith nodeid %s and generated id is %s", conf.ID, nodeid)
						}
						fmt.Printf("Check config at %s with nodeid %s and generated id is %s", fileName, conf.ID, nodeid)
						return nil
					},
				},
			},
		},
		{
			Name:    "node",
			Aliases: []string{"n"},
			Usage:   "test a node with p2p",
			Subcommands: []cli.Command{
				{
					Name:  "start",
					Usage: "start a node",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "c",
							Value: "config.yaml",
							Usage: "default config file",
						},
					},
					Action: func(c *cli.Context) error {
						fileName := c.String("c")
						conf, err := readConfig(fileName)
						if err != nil {
							return err
						}
						node, err := p2p.StartNodeByKey([]byte(conf.PrivateKey))
						if err != nil {
							return err
						}
						p2p.GetListener(node)
						fmt.Println("start node success")
						ch := make(chan os.Signal, 1)
						signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
						<-ch
						return nil
					},
				},
				{
					Name:  "call",
					Usage: "call a node",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "c",
							Usage: "default config file",
						},
					},
					Action: func(c *cli.Context) error {
						fileName := c.String("c")
						var node host.Host
						var err error
						if len(fileName) == 0 {
							node, err = p2p.StartNode()
							if err != nil {
								return err
							}
						} else {
							conf, err := readConfig(fileName)
							if err != nil {
								return err
							}
							node, err = p2p.StartNodeByKey([]byte(conf.PrivateKey))
							if err != nil {
								return err
							}
						}
						p2p.CallRemote(node, "QmXaYoh5YbQWFk4CsZtpn5QtEesLhZaR9B6hLGd5dsh7v2")
						return nil
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig(path string) (*config.Interface, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := config.Interface{}

	// Read in config settings from file.
	err = yaml.Unmarshal(in, &result)
	if err != nil {
		return nil, err
	}

	if result.PrivateKey == "" {
		return nil, errors.New("no private key")
	}
	return &result, nil
}
