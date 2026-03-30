package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	version        string = "0.1.1"
	database       *Database
	addressChannel chan string = make(chan string)
	scanQueue      chan string = make(chan string)
	pingWorkers    int
	scanWorkers    int
	pinging        bool
	scanning       bool
	pinged         int64
	scanned        int64
	valid          int64

	databaseUrl    string
	databaseUser   string
	databasePass   string
	databaseName   string
	logFile        string
	ipFile         string
	startingIp     string
	timeout        int
	maxPingWorkers int
	maxScanWorkers int
	verbose        *bool
)

func main() {
	flag.StringVar(&databaseUrl, "databaseUrl", "localhost:3306", "The URL of the database to store the results in")
	flag.StringVar(&databaseUser, "databaseUser", "root", "The user for the database")
	flag.StringVar(&databasePass, "databasePass", "", "The password for the database")
	flag.StringVar(&databaseName, "databaseName", "openheimer", "The name of the database")
	flag.StringVar(&logFile, "logFile", "openheimer.log", "The file to store the logs in")
	flag.StringVar(&ipFile, "ipFile", "", "The file to extract IP addresses from")
	flag.StringVar(&startingIp, "startingIp", "1.0.0.0", "The IP address to start scanning from")
	flag.IntVar(&timeout, "timeout", 5, "The amount of seconds to wait before timing out")
	flag.IntVar(&maxPingWorkers, "maxPingWorkers", 4000, "The maximum amount of workers to ping IPs")
	flag.IntVar(&maxScanWorkers, "maxScanWorkers", 1000, "The maximum amount of workers to scan IPs")
	verbose = flag.Bool("verbose", false, "Display everything that's happening")
	displayVersion := flag.Bool("version", false, "Display the current version of OpenHeimer")
	flag.Parse()

	if *displayVersion {
		fmt.Printf("OpenHeimer v%v\n", version)
		return
	}

	startTime := time.Now().Unix()
	file, err := os.Create(logFile)
	if err != nil {
		log.Fatalf("Unable to create %v: %v\n", logFile, err.Error())
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	db, err := NewDatabase(databaseUrl, databaseUser, databasePass, databaseName)
	if err != nil {
		fmt.Println(err)
		return
	}
	database = db
	go displayStatus()
	go pingIps()
	go scanIps()
	if ipFile != "" {
		result := readFromFile(ipFile, addressChannel)
		if result == 1 {
			return
		}
	} else {
		result := generateIps(startingIp, addressChannel)
		if result == 1 {
			return
		}
	}

	for pinging || scanning {
		time.Sleep(1 * time.Second)
	}
	log.Printf("Done! Finished in %v seconds. Pinged: %v, Scanned: %v, Valid: %v.\n", time.Now().Unix()-startTime, pinged, scanned, valid)
}

func displayStatus() {
	for {
		time.Sleep(5 * time.Second)
		log.Printf("Pinged: %v, Scanned: %v, Valid: %v\n", pinged, scanned, valid)
	}
}
