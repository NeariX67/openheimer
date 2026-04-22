package main

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
)

func scanIps() {
	scanning = true
	var mutex sync.Mutex
	if *verbose {
		log.Println("Starting IP scanner...")
	}

	for {
		for scanWorkers >= maxScanWorkers {
			time.Sleep(100 * time.Millisecond)
		}

		address := <-scanQueue
		if address == "end" {
			break
		}
		ip := address
		var port uint16 = 25565
		if strings.Contains(address, ":") {
			segments := strings.Split(address, ":")
			ip = segments[0]
			parsedPort, _ := strconv.ParseUint(segments[1], 10, 16)
			port = uint16(parsedPort)
		}
		go scanIp(ip, port, &mutex)
		mutex.Lock()
		scanWorkers++
		scanned++
		mutex.Unlock()
	}

	for scanWorkers > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	scanning = false
}

func scanIp(ip string, port uint16, mutex *sync.Mutex) {
	if *verbose {
		log.Printf("Scanning %v:%v...\n", ip, port)
	}
	startTime := time.Now()
	response, err := mcpinger.New(ip, port, mcpinger.WithTimeout(time.Second*time.Duration(timeout))).Ping()
	if err != nil {
		if *verbose {
			log.Printf("Unable to scan %v:%v: %v\n", ip, port, err.Error())
		}
		mutex.Lock()
		scanWorkers--
		mutex.Unlock()
		return
	}
	ping := time.Since(startTime).Milliseconds()

	entry := &ServerEntry{}
	entry.FromServerInfo(ip, port, response)

	loc, err := fetchIpLocation(ip)
	if err == nil {
		entry.CountryCode = loc.CountryCode2
	}
	entry.Ping = int32(ping)

	log.Printf("Found Minecraft server at %v:%v (Ping: %d ms)\n", ip, port, ping)
	err = database.Write(*entry)
	if err != nil {
		log.Printf("Unable to write to database: %v\n", err.Error())
	}

	mutex.Lock()
	scanWorkers--
	valid++
	mutex.Unlock()
}
