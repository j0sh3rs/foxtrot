package main

import (
	"bufio"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// customUserAgent allows setting a custom user agent, defaulting to the specified one.
var customUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0"

// loadWebsites loads the list of websites from the specified file.
func loadWebsites(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var websites []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		websites = append(websites, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return websites, nil
}

// sendRequest sends an HTTP request to the given website with the custom user agent.
func sendRequest(website string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirects
		},
	}

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

func main() {
	rand.Seed(time.Now().UnixNano())
	websites, err := loadWebsites("/data/sites.txt")
	if err != nil {
		log.Fatalf("Error loading websites: %v", err)
	}

	var wg sync.WaitGroup
	for _, website := range websites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			for {
				sendRequest(site)
				time.Sleep(time.Duration(rand.Intn(45-1)+1) * time.Second)
			}
		}(website)
	}

	wg.Wait()
}
