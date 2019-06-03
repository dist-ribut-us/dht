package main

import (
	"github.com/dist-ribut-us/dht/dhtnetwork/sim"
	"time"
)

func main() {
	net := sim.New()
	net.AddNodes(200)

	net.Run()
	time.Sleep(time.Minute * 10)
}
