package main

import (
	"flag"
	"fmt"
	"github.com/middlehut/btclisten/pkg/btc"
	"log"
	"net"
	"os"
)

var seedNodeIP string

func main(){
	flag.StringVar(&seedNodeIP, "seednode", "", "bitcoin node IP address to connect to")
	flag.Parse()

	if seedNodeIP == "" {
		flag.Usage()
		os.Exit(0)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:8333", seedNodeIP))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("connected to %s\n", seedNodeIP)

	done := make(chan bool)
	listener := btc.NewListener(conn, done)
	listener.Start()
	<-done

	fmt.Println("disconnected")
}

func usage() {
	fmt.Println("")
}