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
 - go1.13.5


### How to use it ?

To use it, simply install it with the following command
```
go get github.com/reversTeam/go-ms
```
This command will allow you to add the framework directly to your `$ GOPATH`, that is to say that you do not have to` git clone` and that you can simply import it into your project.
You can look at the example files which will allow you to deploy the different servers:
 - grpc
 - http
 - gateway = grpc + http

We will take the example of the file which makes it possible to make the gateway, because this one has the merit of launching the two servers (grpc + http)
```golang
package main

import (
	"flag"
	"github.com/reversTeam/go-ms/core"
	"github.com/reversTeam/go-ms/services/goms"
	"github.com/reversTeam/go-ms/services/child"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

const (
	// Default flag values for GRPC server
	GRPC_DEFAULT_HOST = "127.0.0.1"
	GRPC_DEFAULT_PORT = 42001

	// Default flag values for http server
	HTTP_DEFAULT_HOST = "127.0.0.1"
	HTTP_DEFAULT_PORT = 8080
)

var (
	// flags for Grpc server
	grpcHost = flag.String("grpc-host", GRPC_DEFAULT_HOST, "Grpc listening host")
	grpcPort = flag.Int("grpc-port", GRPC_DEFAULT_PORT, "Grpc listening port")

	// flags for http server
	httpHost = flag.String("http-host", HTTP_DEFAULT_HOST, "http gateway host")
	httpPort = flag.Int("http-port", HTTP_DEFAULT_PORT, "http gateway port")
)

func main() {
	// Instantiate context in background
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Parse flags
	flag.Parse()

	// Create a gateway configuration
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// setup servers
	grpcServer := core.NewGoMsGrpcServer(ctx, *grpcHost, *grpcPort, opts)
	httpServer := core.NewGoMsHttpServer(ctx, *httpHost, *httpPort, grpcServer)

	// setup services
	gomsService := goms.NewService("goms")
	childService := child.NewService("child")

	// Register service to the grpc server
	grpcServer.AddService(gomsService)
	grpcServer.AddService(childService)

	// Register service to the http server
	httpServer.AddService(gomsService)
	httpServer.AddService(childService)

	// Graceful stop servers
	core.AddServerGracefulStop(grpcServer)
	core.AddServerGracefulStop(httpServer)
	// Catch ctrl + c
	done := core.CatchStopSignals()

	// Start Grpc Server
	err := grpcServer.Start()
	if err != nil {
		log.Fatal("An error occured, the grpc server can be running", err)
	}
	// Start Http Server
	err = httpServer.Start()
	if err != nil {
		log.Fatal("An error occured, the http server can be running", err)
	}

	<-done
}
```

If we break down the code we have above we can see that we have different phases:
 - Import of libraries which are necessary to operate your hand
```golang
import (
	"flag"
	"github.com/reversTeam/go-ms/core"
	"github.com/reversTeam/go-ms/services/goms"   // Only for example
	"github.com/reversTeam/go-ms/services/child"  // Only for example
	// "github.com/yoursName/go-ms-service-what-you-want"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)
 ```
 - Initialization of the constants for the default values ​​of the flags, readability questions
 ```golang
 const (
	// Default flag values for GRPC server
	GRPC_DEFAULT_HOST = "127.0.0.1"
	GRPC_DEFAULT_PORT = 42001

	// Default flag values for http server
	HTTP_DEFAULT_HOST = "127.0.0.1"
	HTTP_DEFAULT_PORT = 8080
)
 ```
 - Initialization of program flags in global variables, questionable but extremely readable
```golang
var (
	// flags for Grpc server
	grpcHost = flag.String("grpc-host", GRPC_DEFAULT_HOST, "Grpc listening host")
	grpcPort = flag.Int("grpc-port", GRPC_DEFAULT_PORT, "Grpc listening port")

	// flags for http server
	httpHost = flag.String("http-host", HTTP_DEFAULT_HOST, "http gateway host")
	httpPort = flag.Int("http-port", HTTP_DEFAULT_PORT, "http gateway port")
)
```
 - Initialization of grpc and http servers
```golang
// Instantiate context in background
ctx := context.Background()
ctx, cancel := context.WithCancel(ctx)
defer cancel()

// Parse flags
flag.Parse()

// Create a gateway configuration
opts := []grpc.DialOption{
	grpc.WithInsecure(),
}

// setup servers
grpcServer := core.NewGoMsGrpcServer(ctx, *grpcHost, *grpcPort, opts)
httpServer := core.NewGoMsHttpServer(ctx, *httpHost, *httpPort, grpcServer)
```
 - Service initialization
   If you create your own modules try to respect this name as well as possible for your repositories, I would try afterwards to make a service manager that everyone can offer their own services.
```golang
// setup services

gomsService := goms.NewService("goms")    // import "github.com/reversTeam/go-ms/services/goms"
childService := child.NewService("child") // import "github.com/reversTeam/go-ms/services/child"
whatYouWantService := whatYouWant.NewService("what-you-want") // import "github.com/yoursName/go-ms-service-what-you-want"
```
 - Ajout des services sur les différents serveurs
```golang
// Register service to the grpc server
grpcServer.AddService(gomsService)
grpcServer.AddService(childService)

// Register service to the http server
httpServer.AddService(gomsService)
httpServer.AddService(childService)
```
 - Ajout des signaux pour couper les services
```golang
// Graceful stop servers
core.AddServerGracefulStop(grpcServer)
core.AddServerGracefulStop(httpServer)
// Catch ctrl + c
done := core.CatchStopSignals()
```
 - Launch of different servers
 If you want to start only one of the two servers, delete the code that starts the one you don't want. In case you launch an http server, you will have to give it the configuration of a functional grpc server.
```golang
// Start Grpc Server
err := grpcServer.Start()
if err != nil {
	log.Fatal("An error occured, the grpc server can be running", err)
}
// Start Http Server
err = httpServer.Start()
if err != nil {
	log.Fatal("An error occured, the http server can be running", err)
}
```
 - We are waiting for the signal telling us to finish the services, in the case of a ctrl + c for example
```golang
<-done
```

### Credits

 - go-micro
 - golang
 - protoc