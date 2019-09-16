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
	seedNodeIP   string
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

	fmt.Printf("connected to %s\n", conn.RemoteAddr().String())

	quit := make(chan struct{})
	listener, err := btc.NewListener(conn)
	if err != nil {
		log.Fatal(err)
	}
	go listener.Start(quit)

	// handle CTRL+C
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-quit:
		break
	case <-c:
		listener.Stop()
		<-quit
	}

	fmt.Println("disconnected")
}
