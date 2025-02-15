package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/crepehat/speedtest-cli/speedtest"
)

func version() {
	fmt.Print(speedtest.Version)
}

func usage() {
	fmt.Fprint(os.Stderr, "Command line interface for testing internet bandwidth using speedtest.net.\n\n")
	flag.PrintDefaults()
}

func main() {
	opts := speedtest.ParseOpts()

	switch {
	case opts.Help:
		usage()
		return
	case opts.Version:
		version()
		return
	}

	client := speedtest.NewClient(opts)

	if opts.List {
		servers, err := client.AllServers()
		if err != nil {
			log.Fatalf("Failed to load server list: %v\n", err)
		}
		fmt.Println(servers)
		return
	}

	config, err := client.Config()
	if err != nil {
		log.Fatal(err)
	}

	client.Log("Testing from %s (%s)...\n", config.Client.ISP, config.Client.IP)

	ticker := time.NewTicker(1 * time.Minute)

	for ; true; <-ticker.C {
		server := selectServer(opts, client)
		downloadSpeed := server.DownloadSpeed()
		uploadSpeed := server.UploadSpeed()
		// ping is ms, speeds are megabytes
		log.Printf("ping: %d;down:%.2f;up:%.2f", server.Latency/time.Millisecond, float64(downloadSpeed)/(1<<20), float64(uploadSpeed)/(1<<20))
	}
}

// func reportSpeed(opts *speedtest.Opts, prefix string, speed int) {
// 	if opts.SpeedInBytes {
// 		fmt.Printf("%s: %.2f MiB/s\n", prefix, float64(speed)/(1<<20))
// 	} else {
// 		fmt.Printf("%s: %.2f Mib/s\n", prefix, float64(speed)/(1<<17))
// 	}
// }

func selectServer(opts *speedtest.Opts, client speedtest.Client) (selected *speedtest.Server) {
	if opts.Server != 0 {
		servers, err := client.AllServers()
		if err != nil {
			log.Fatal("Failed to load server list: %v\n", err)
			return nil
		}
		selected = servers.Find(opts.Server)
		if selected == nil {
			log.Fatalf("Server not found: %d\n", opts.Server)
			return nil
		}
		selected.MeasureLatency(speedtest.DefaultLatencyMeasureTimes, speedtest.DefaultErrorLatency)
	} else {
		servers, err := client.ClosestServers()
		if err != nil {
			log.Fatal("Failed to load server list: %v\n", err)
			return nil
		}
		selected = servers.MeasureLatencies(
			speedtest.DefaultLatencyMeasureTimes,
			speedtest.DefaultErrorLatency).First()
	}

	if !opts.Quiet {
		client.Log("Hosted by %s (%s) [%.2f km]: %d ms\n",
			selected.Sponsor,
			selected.Name,
			selected.Distance,
			selected.Latency/time.Millisecond)
	}

	return selected
}
