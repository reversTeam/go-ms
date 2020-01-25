# GoMs - Framework for distributed protobuf micro services

Ce framework go est encore en dévelopemment et cette mention sera retirer lorsque j'estimerais que le projet a une maturitée suffisante pour commencer à être exploité par d'autre personne.

### Pourquoi ce projet ?

Dernièrement j'ai fais le tours de pas mal de techno et je dois reconnaitre que ces dernières années nous faisons face à pas mal de nouvelles technologies qui sont trop peu exploité par les entreprises, elles estiment bien souvent que ce n'est qu'un phénomène de hipe.
Pour ma part je reste convaincus que les technologies sont une nouvelles fois bouleversé par l'arrivé de nouvelles technologies et de nouveau mode de consommation.
De plus, certaines choses me laisse à penser qu'il ne faut rien prendre pour acquis et qu'avec le temps des connaissances ce perdent, donc autant faire ce projet proprement, et non pas comme de nombreux labs qui se sont perdus sur des disques dur et qui n'ont jamais aboutis.
Cela permet aussi au gens qui auraient envie de découvrir une ou plusieurs technologie présente dans ce projet.

Je vous communique une idée, puisse t-elle être utilisé à bon escient.

### Qu'est ce que c'est ?

C'est un framework ecris en go, qui vous permettra de construire une stack micro service distribuée avec un format d'échange en protobuf très simplement.
GoMs vous permet de vous concentrer principalement sur vos services, le fait de register les rpc au serveur grpc ou le serveur grpc au serveur http est totalement abstrait. Vous n'avez qu'a register un service à un serveur.

L'objectif c'est de rendre le service totalement autonome de ses handler, il ne sont que des couches d'expositions. Si vous avez besoin des ressources qui se trouve dans un autre module, vous passerais directement par un call grpc protobuf, c'est a dire que :
 - Vous allez intérroger un autre serveur avec une socket qui est déjà connecter avant même que vous ayez demander de faire le call, contrairement à HTTP
 - Vous n'avez pas la couche HEADER du protocole HTTP qui dans beaucoup de cas sont plus gros que les données qu'elles transporte
 - Vous avez un flux binaire, ce n'est pas un format de serialization comme JSON, c'est a dire:
 	- Qui plante au runtime, une typo, bad format, etc..
 	- Flux non biaire, ajout de caractère de contexte `[]{}()"",:<tab><space><etc...>`
 - Vous savez à la compilation si ça va fontionner ou non
 - Vous être libre de déployer vos services comme vous le voulez:
    - sur un serveur
    - sur plusieurs
    - seulement le gppc
    - seulement le http

### Dépendances

Ce projet demande un minimum de packets afin de garantir son fonctionnement. Veuillez installer les librairies suivantes:
 - libprotoc 3.11.2
 - go1.13.5


### Comment l'utiliser ?

Pour l'utliser, il vous suffira de l'installer avec la commande suivante
```
go get github.com/reversTeam/go-ms
```
Cette commande vous permettra d'ajouter le framework directement dans votre `$GOPATH`, c'est à dire que vous n'avez pas à le `git clone` et que vous pouvez simplement l'importer dans votre projet.
Vous pouvez regarder les fichiers d'exemple qui vous permettrons de déployer les différents servers:
 - grpc
 - http
 - gateway = grpc + http

Nous prendrons l'exemple du fichier qui permet de faire la gateway, car celle-ci à le mérite de lancer les deux serveurs (grpc + http)
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

Si nous décomposons le code que nous avons plus haut nous pouvons voir que nous avons différentes phases :
 - Import des librairies qui sont nécessaire a faire fonctionner votre main
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
 - Initialisation des constantes pour les default values des flags, questions de lisibilité
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
 - Initialisation des flags du programmes dans des varibales globales, discutable mais extrèmement lisible
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
 - Initialisation des serveurs grpc et http
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
 - Initialisation des services
   Si vous creer vos propre modules essayez de respecter au mieux ce nom pour vos repository, j'essaierais par la suite de faire un service manager que chacun puisse proposer ses propres services.
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
 - Lancement des différents serveurs
 Si vous voulez lancer seulement un des deux serveur supprimer le code qui lance celui que vous ne voulez pas. Dans le cas ou vous lancer un serveur http, il faudra lui donner les configuration d'un serveur grpc fonctionnelle.
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
 - On attend le signal qui nous dis de terminer les services, dans le cas d'un ctrl + c par example
```golang
<-done
```

### Credits

 - go-micro
 - golang
 - protoc