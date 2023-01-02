package main

import (
	"os"
	"os/user"
	"strconv"

	dknet "github.com/docker/go-plugins-helpers/network"
	i2p "github.com/eyedeekay/docker-i2p-plugin/i2p"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	version = "0.2"
)

func main() {

	var flagDebug = cli.BoolFlag{
		Name:  "debug, d",
		Usage: "enable debugging",
	}
	app := cli.NewApp()
	app.Name = "don"
	app.Usage = "Docker Open vSwitch Networking"
	app.Version = version
	app.Flags = []cli.Flag{
		flagDebug,
	}
	app.Action = Run
	app.Run(os.Args)
}

func lookupGroup(group string) (gid int, err error) {
	g, err := user.LookupGroup(group)
	if err != nil {
		return 0, err
	}
	gid, err = strconv.Atoi(g.Gid)
	if err != nil {
		return 0, err
	}
	return
}

// Run initializes the driver
func Run(ctx *cli.Context) {
	if ctx.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	d, err := i2p.NewDriver()
	if err != nil {
		panic(err)
	}
	gid, err := lookupGroup("i2psvc")
	if err != nil {
		panic(err)
	}
	h := dknet.NewHandler(d)
	h.ServeUnix("root", gid)
}
