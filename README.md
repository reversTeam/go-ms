# GoMs - Framework for distributed protobuf micro services

This go framework is still under development and this mention will be withdrawn when I consider that the project has sufficient maturity to start being exploited by other people.

### Why this project?

Lately I have been around a lot of techno and I must admit that in recent years we face a lot of new technologies which are too little exploited by companies, they often believe that it is only a hipe phenomenon.
For my part I remain convinced that technologies are once again upset by the arrival of new technologies and new mode of consumption.
In addition, some things leave me to think that nothing should be taken for granted and that with time knowledge it loses, so much to do this project properly, and not like many labs that got lost on discs hard and never succeeded.
This also allows people who would like to discover one or more technologies present in this project.

I am communicating an idea to you, may it be used wisely.

### What is that ?

It is a framework written in go, which will allow you to build a micro service stack distributed with an exchange format in protobuf very simply.
GoMs allows you to focus mainly on your services, registering the rpc to the grpc server or the grpc server to the http server is completely abstract. You only have to register a service to a server.

The objective is to make the service completely independent of its handlers, there are only layers of exhibitions. If you need the resources which is in another module, you would pass directly by a call grpc protobuf, ie that:
 - You are going to query another server with a socket that is already connected before you even ask to make the call, unlike HTTP
 - You do not have the HEADER layer of the HTTP protocol which in many cases is larger than the data it carries
 - You have a bit stream, it is not a serialization format like JSON, that is to say:
 	- Who crashes in runtime, a typo, bad format, etc.
 	- Non-binary stream, addition of context character `[]{}()"",:<tab><space><etc...>`
 - You know at compilation whether it will work or not
 - You are free to deploy your services as you wish:
    - on a server
    - on multiple servers
    - only the gppc server
    - only http server

### Dependencies

This project requires a minimum of packets to guarantee its functioning. Please install the following libraries:
 - libprotoc 3.11.2
 - go1.20


### How to use it ?

To use it, simply install it with the following command
```
go get github.com/reversTeam/go-ms
go mod tidy
```
This command will allow you to add the framework directly to your `$ GOPATH`, that is to say that you do not have to` git clone` and that you can simply import it into your project.
You can look at the example files which will allow you to deploy the different servers:
 - grpc
 - http
 - gateway = grpc + http

We will take the example of the file which makes it possible to make the gateway, because this one has the merit of launching the two servers (grpc + http)

1. Create a config file in `./config/config.yml`
```yaml
grpc:
  host: "127.0.0.1"
  port: 42001
http:
  host: "127.0.0.1"
  port: 8080
exporter:
  host: "127.0.0.1"
  port: 4242
  path: "/metrics"
  interval: 1
services:
  goms:
    grpc: true
    http: true
    config:
      database:
        host: 127.0.0.1
        port: 3306
  child:
    grpc: true
    http: true
    config:
      database:
        host: 127.0.0.1
        port: 5432
```

2. Create the main.go
```golang
package main

import (
	"flag"
	"log"

	"github.com/reversTeam/go-ms/core"
	"github.com/reversTeam/go-ms/services/child"
	"github.com/reversTeam/go-ms/services/goms"
)

const (
	GO_MS_CONFIG_FILEPATH = "./config/config.yml"
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
```

### Run server :
```bash
go run main.go
```

With an other config file:
```bash
go run main.go -config other/path/to/config.yml
```