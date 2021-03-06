package main

import (
	"fmt"
	"os"

	"coding.pickflames.com/pickflames/framework/utils/configurator"
	"coding.pickflames.com/pickflames/framework/utils/log"
	"github.com/urfave/cli"
)

type Config struct {
	RESTConfig   RESTConfig   `conf:"rest"`
	DockerConfig DockerConfig `conf:"docker"`
}

var version string
var build string

func main() {

	log.Infof("version(%s) build(%s)", version, build)
	app := cli.NewApp()
	app.Name = "monitor"
	app.Usage = "start monitor"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		//
		// rest
		//
		cli.StringFlag{
			Name:   "rest-address",
			EnvVar: "REST_ADDRESS",
			Value:  "127.0.0.1:9001",
		},
		cli.StringFlag{
			Name:   "rest-access-token",
			EnvVar: "REST_ACCESS_TOKEN",
			Value:  "test",
		},

		//
		// docker
		//
		cli.StringFlag{
			Name:   "docker-registry-file",
			EnvVar: "DOCKER_REGISTRY_FILE",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	conf := &Config{}
	if err := configurator.New(c).Configure(conf); err != nil {
		log.Errorf("configure failed: %v", err)
		return err
	}

	dc, err := NewDockerClient(&conf.DockerConfig)
	if err != nil {
		log.Errorf("create docker client failed: %v", err)
		return err
	}

	srv, err := NewServer(&conf.RESTConfig)
	if err != nil {
		log.Errorf("create server failed: %v", err)
		return err
	}

	if err := srv.InitRoutes(dc); err != nil {
		log.Errorf("server init routes failed: %v", err)
		return err
	}

	if err := srv.Run(); err != nil {
		log.Errorf("server run failed: %v", err)
		return err
	}
	return nil
}
