package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/penkovski/btclisten/pkg/btc"
)

var (
	// seedNodeIP specifies the initial IP to connect to
	seedNodeIP string
	// seedNodePort specifies the initial PORT to connect to
	seedNodePort string
)

func main() {
	flag.StringVar(&seedNodeIP, "seedip", "", "bitcoin node IP address to connect to")
	flag.StringVar(&seedNodePort, "seedport", "8333", "bitcoin node Port to connect to")
	flag.Parse()

	if seedNodeIP == "" {
		flag.Usage()
		os.Exit(0)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", seedNodeIP, seedNodePort))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("connected to %s\n", conn.RemoteAddr().String())

	done := make(chan struct{})
	listener, err := btc.NewListener(conn)
	if err != nil {
		log.Fatal(err)
	}
	go listener.Run(done)

	// handle CTRL+C
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-done:
		break
	case <-sigch:
		listener.Stop()
		<-done
	}

	log.Println("disconnected")
}
