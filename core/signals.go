package core

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	servers []GoMsServerGracefulStopableInterface
)

func AddServerGracefulStop(server GoMsServerGracefulStopableInterface) {
	servers = append(servers, server)
}

// Catch SIG_TERM and exit propely
func CatchStopSignals() (done chan bool) {
	done = make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer log.Println("[SYSTEM]: System is ready for catch exit's signals, To exit press CTRL+C")

	go func() {
		sig := <-sigs
		log.Println("[SYSTEM]: Signal catch:", sig)
		for _, server := range servers {
			err := server.GracefulStop()
			if err != nil {
				log.Println("Server can't GracefulStop", err)
			}
		}
		done <- true
	}()
	return
}
