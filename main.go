package main

import (
	"flag"
	"strings"

	"github.com/UnplugCharger/shreder/shreder"
	"github.com/rs/zerolog/log"
)

var port string
var peers string

func main() {
	flag.StringVar(&port, "port", ":8060", "Port to run the cache server on")
	flag.StringVar(&peers, "peers", "", "Comma separated list of peers to replicate to")

	flag.Parse()

	var peerList []string
	if peers != "" {
		peerList = strings.Split(peers, ",")
		// Filter out empty strings
		var filteredPeers []string
		for _, peer := range peerList {
			if strings.TrimSpace(peer) != "" {
				filteredPeers = append(filteredPeers, strings.TrimSpace(peer))
			}
		}
		peerList = filteredPeers
	}
	
	log.Printf("Starting server on port %s with peers: %v", port, peerList)
	cache := shreder.NewCacheServer(peerList, port)

	err := cache.Start(port)
	if err != nil {
		panic(err)
	}
}
