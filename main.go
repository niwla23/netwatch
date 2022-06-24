package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-ping/ping"
	"github.com/j-keck/arping"
	"github.com/korylprince/ipnetgen"
)

type HostAnalysis struct {
	ping_stats *ping.Statistics
	hostnames  *[]string
	hw_adress  *net.HardwareAddr
	up         bool
}

func smallerThanActualMaxCount(count int, maxCount int) bool {
	if maxCount == 0 {
		return true
	} else {
		return count < maxCount
	}
}

func analyze_host(address net.IPAddr, timeout int, maxCount int, stats chan HostAnalysis, wg *sync.WaitGroup) {
	defer wg.Done()

	// fmt.Println("pinged", address)
	for count := 0; smallerThanActualMaxCount(count, maxCount); count++ {
		pinger := ping.New("")
		pinger.SetIPAddr(&address)

		pinger.Count = 1
		pinger.Timeout = time.Duration(timeout) * time.Millisecond

		pinger.Run()

		hostnames, _ := net.LookupAddr(address.IP.String())
		analysis := HostAnalysis{ping_stats: pinger.Statistics(), hostnames: &hostnames, up: true}

		if hwAddr, _, err := arping.Ping(address.IP); err == nil {
			analysis.hw_adress = &hwAddr
		}

		if analysis.ping_stats.PacketLoss != 100 && !math.IsNaN(analysis.ping_stats.PacketLoss) {
			stats <- analysis
			break
		}
	}
	stats <- HostAnalysis{up: false}
}

func main() {
	var subnet string
	var maxCount int
	var timeout int
	flag.StringVar(&subnet, "subnet", "REQUIRED", "The subnet to scan, e.g. 192.168.178.0/24")
	flag.IntVar(&maxCount, "max-count", 0, "The maximal amount of pings per host")
	flag.IntVar(&timeout, "timeout", 5000, "Timeout for ICMP pings")
	flag.Parse()

	if subnet == "REQUIRED" {
		color.Red("The --subnet argument is required")
		os.Exit(1)
	}

	wg := new(sync.WaitGroup)

	stats := make(chan HostAnalysis)

	gen, err := ipnetgen.New(subnet)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	for ip := gen.Next(); ip != nil; ip = gen.Next() {
		wg.Add(1)
		go analyze_host(net.IPAddr{IP: ip, Zone: ""}, timeout, maxCount, stats, wg)
	}

	ip_color := color.New(color.FgBlue)
	hostname_color := color.New(color.FgGreen)
	mac_color := color.New(color.FgYellow)

	gen2, _ := ipnetgen.New(subnet)
	for ip := gen2.Next(); ip != nil; ip = gen2.Next() {
		stats_received := <-stats
		if stats_received.up {
			ip_color.Print(stats_received.ping_stats.Addr)
			fmt.Print(" ")
			hostname_color.Print(*stats_received.hostnames)
			if stats_received.hw_adress != nil {
				fmt.Print(" ")
				mac_color.Print(stats_received.hw_adress)
			}
			fmt.Println()
		}
	}

}
