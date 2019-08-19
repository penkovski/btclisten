package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

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

	fmt.Printf("connected to %s\n", seedNodeIP)

	quit := make(chan struct{})
	listener, err := btc.NewListener(conn)
	if err != nil {
		log.Fatal(err)
	}
	go listener.Start(quit)

	// TODO(penkovski): CTRL-C handler

	<-quit
	fmt.Println("disconnected")
}
