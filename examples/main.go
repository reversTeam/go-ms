package main

import (
	"flag"
	"log"

	"github.com/reversTeam/go-ms/core"
	"github.com/reversTeam/go-ms/services/child"
	"github.com/reversTeam/go-ms/services/goms"
)

const (
	GO_MS_CONFIG_FILEPATH = "./gateway/config/config.yml"
)

var (
	configFilePath = flag.String("config", GO_MS_CONFIG_FILEPATH, "yaml config filepath")
)

var (
	goMsServices = map[string]func(string, core.ServiceConfig) core.GoMsServiceInterface{
		"goms": core.RegisterServiceMap(func(name string, config core.ServiceConfig) core.GoMsServiceInterface {
			return goms.NewService(name, config)
		}),
		"child": core.RegisterServiceMap(func(name string, config core.ServiceConfig) core.GoMsServiceInterface {
			return child.NewService(name, config)
		}),
	}
)

func main() {
	flag.Parse()
	config, err := core.NewConfig(*configFilePath)
	if err != nil {
		log.Panic(err)
	}

	app := core.NewApplication(config, goMsServices)
	app.Start()
}
