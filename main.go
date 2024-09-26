package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"io/ioutil"
	"sync"
	"path/filepath"
)

type Site struct {
	URL    string `json:"site"`
	Protocol string `json:"protocol"`
	Port   int    `json:"port"`
}

func main() {
	fmt.Println("PROGRAM STARTED")

	// Create a ticker for periodic checks
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			// Read the data.json file
			data, err := ioutil.ReadFile("data.json")
			if err != nil {
				log.Fatal(err)
			}
	
			var sites []Site
			err = json.Unmarshal(data, &sites)
			if err != nil {
				log.Fatal(err)
			}

			timestamp := time.Now().Unix()

			directory := "logs/"

			errCreateDir := os.MkdirAll(directory, os.ModePerm)
			if errCreateDir != nil {
				fmt.Errorf("Error creating directory: %w", err)
			}

			filename := filepath.Join(directory, fmt.Sprintf("sites_logs_%d.log", timestamp))

			// Create a log file for the current timestamp
			logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
					log.Fatal(err)
			}
			defer logFile.Close()

			var wg sync.WaitGroup

			// Iterate through sites and perform checks
			fmt.Println("Checking Site Availability now...")
			for _, site := range sites {
				wg.Add(1) 
                go func(s Site) {
                    defer wg.Done()
                    checkSite(s, logFile)
                }(site)
			}

			wg.Wait() 
		}
	}
}

func checkSite(site Site, logFile *os.File) {

	switch site.Protocol {
		case "HTTP":
			// Create an HTTP client
			client := &http.Client{}

			req, err := http.NewRequest("GET", site.URL, nil)
			if err != nil {
				log.Println("Error creating request:", err)
				return
			}
		
			startTime := time.Now()

			// Send the request
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error sending request:", err)
				return
			}
			defer resp.Body.Close()

			elapsedTime := time.Since(startTime)
			log.Printf("%s %s %s %d %s", time.Now().Format("2006-01-02 15:04:05"), site.URL, site.Protocol, resp.StatusCode, elapsedTime)
			logFile.WriteString(fmt.Sprintf("%s %s %s %d %s\n", time.Now().Format("2006-01-02 15:04:05"), site.URL, site.Protocol, resp.StatusCode, elapsedTime))
		case "TCP":
			startTime := time.Now()

			// Perform TCP port check
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", site.URL, site.Port))
			if err != nil {
					log.Println(err)
					return
			}
			defer conn.Close()

			// Log TCP connection success and response time
			elapsedTime := time.Since(startTime)
			log.Printf("%s %s %s TCP port open %s", time.Now().Format("2006-01-02 15:04:05"), site.URL, site.Protocol, elapsedTime)
			logFile.WriteString(fmt.Sprintf("%s %s %s OPEN %s\n", time.Now().Format("2006-01-02 15:04:05"), site.URL, site.Protocol, elapsedTime))
		default:
			// Handle unsupported protocols or custom checks
			log.Println("Unsupported protocol:", site.Protocol)
	}
}