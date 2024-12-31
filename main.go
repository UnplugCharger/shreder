package main

import (
	"flag"
	"github.com/UnplugCharger/shreder/shreder"
	"strings"
)

var port string
var peers string

func main() {
	flag.StringVar(&port, "port", ":8083", "Port to run the cache server on")
	flag.StringVar(&peers, "peers", "", "Comma separated list of peers to replicate to")

	flag.Parse()

	peerList := strings.Split(peers, ",")

	cache := shreder.NewCacheServer(peerList)

	err := cache.Start(port)
	if err != nil {
		panic(err)
	}

}
