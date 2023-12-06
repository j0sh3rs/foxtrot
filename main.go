package main

import (
	"encoding/csv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var customUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0"

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

		website := record[0]
		if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			website = "https://" + website
		}
		websites = append(websites, website)
	}

	return websites, nil
}

func selectRandomWebsites(websites []string, count int) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(websites), func(i, j int) {
		websites[i], websites[j] = websites[j], websites[i]
	})
	if count > len(websites) {
		count = len(websites)
	}
	return websites[:count]
}

func main() {
	allWebsites, err := downloadWebsites("https://analytics.usa.gov/data/live/sites.csv")
	if err != nil {
		log.Fatalf("Error downloading websites: %v", err)
	}

	client := &http.Client{}
	var wg sync.WaitGroup
	quit := make(chan struct{})
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-ticker.C:
				// Stop existing goroutines
				close(quit)
				quit = make(chan struct{})
				wg = sync.WaitGroup{}

				// Select 10 random websites
				selectedWebsites := selectRandomWebsites(allWebsites, 10)

				// Start new goroutines for the new set of websites
				for _, website := range selectedWebsites {
					wg.Add(1)
					go func(site string) {
						defer wg.Done()
						for {
							select {
							case <-quit:
								return
							default:
								sendRequest(client, site)
								time.Sleep(time.Duration(rand.Intn(45-1)+1) * time.Second)
							}
						}
					}(website)
				}
			}
		}
	}()

	wg.Wait()
	ticker.Stop()
}
