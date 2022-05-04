package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type HostAnalysis struct {
	ping_stats ping.Statistics
	hostnames  []string
}

func analyze_host(address net.IPAddr, stats chan HostAnalysis, wg *sync.WaitGroup) {
	defer wg.Done()
	pinger := ping.New("")
	pinger.SetIPAddr(&address)

	pinger.Count = 1
	pinger.Timeout = 5 * time.Second

	err := pinger.Run()
	if err != nil {
		panic(err)
	}

	hostnames, _ := net.LookupAddr(address.IP.String())

	analysis := HostAnalysis{ping_stats: *pinger.Statistics(), hostnames: hostnames}
	stats <- analysis
}

func main() {
	wg := new(sync.WaitGroup)

	stats := make(chan HostAnalysis)

	for i := 0; i < 255; i++ {
		ipv4 := net.IPv4(192, 168, 178, byte(i))
		wg.Add(1)
		go analyze_host(net.IPAddr{IP: ipv4, Zone: ""}, stats, wg)
	}

	for i := 0; i < 100; i++ {
		stats_received := <-stats
		if stats_received.ping_stats.PacketLoss != 100 {
			fmt.Println(stats_received.ping_stats.Addr, stats_received.hostnames)
		}
	}

}
