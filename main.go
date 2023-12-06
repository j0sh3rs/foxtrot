package main

import (
	"encoding/csv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var customUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0"

// sendRequest sends an HTTP request to the given website with the custom user agent.
func sendRequest(client *http.Client, website string) {
	req, err := http.NewRequest("GET", website, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", website, err)
		return
	}

	req.Header.Set("User-Agent", customUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to %s: %v", website, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Request to %s completed with status code: %d", website, resp.StatusCode)
}

// downloadWebsites downloads and parses the list of websites from a CSV file at the given URL.
func downloadWebsites(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	var websites []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Assuming the website URL is in the first column
		websites = append(websites, record[0])
	}

	return websites, nil
}

func main() {
	websites, err := downloadWebsites("https://analytics.usa.gov/data/live/sites.csv")
	if err != nil {
		log.Fatalf("Error downloading websites: %v", err)
	}

	client := &http.Client{}

	var wg sync.WaitGroup
	for _, website := range websites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			for {
				sendRequest(client, site)
				time.Sleep(time.Duration(rand.Intn(45-1)+1) * time.Second)
			}
		}(website)
	}

	wg.Wait()
}
