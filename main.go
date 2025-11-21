package main

import (
	"coding.pickflames.com/pickflames/framework/cli"
	"coding.pickflames.com/pickflames/framework/utils/log"
)

type Config struct {
	RESTConfig   RESTConfig   `conf:"rest"`
	DockerConfig DockerConfig `conf:"docker"`
}

var version = "test-version"
var build = "test-build"
var name = "monitor"

func main() {
	app := cli.NewApp(name, version, build)
	app.SetUsage(name)

	conf := &Config{}
	app.BindConfig(conf)
	app.AddCommand("start", "start monitor",
		func() error {
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
		})

	app.Run()
}
