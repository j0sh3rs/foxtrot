// Package foxtrot is meant to help bump the numbers of the Firefox User-Agent against US Gov Websites
package main

import (
	"context"
	"encoding/csv"
	"expvar"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/dnscache"
	"github.com/spf13/cobra"
)

var (
	concurrency        int
	delay              int
	userAgent          string
	totalRequests      int64
	successfulRequests map[string]int64
	failedRequests     map[string]int64
	mu                 sync.Mutex
	location           *time.Location
)

func init() {

	successfulRequests = make(map[string]int64)
	failedRequests = make(map[string]int64)
	setTimeZone()
}

func setTimeZone() {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC" // Default to UTC if no timezone is set
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Fatalf("Failed to load timezone %s: %v", tz, err)
	}
	location = loc
}

func sendRequest(client *http.Client, website string) {
	req, err := http.NewRequest("GET", website, nil)
	if err != nil {
		log.Printf("Error creating request for %s: %v", website, err)
		incrementFailedRequests(website)
		return
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to %s: %v", website, err)
		incrementFailedRequests(website)
		return
	}
	defer resp.Body.Close()

	incrementTotalRequests()
	if resp.StatusCode == http.StatusOK {
		incrementSuccessfulRequests(website)
	} else {
		incrementFailedRequests(website)
	}

	log.Printf("Request to %s completed with status code: %d", website, resp.StatusCode)
}

func incrementTotalRequests() {
	mu.Lock()
	totalRequests++
	mu.Unlock()
}

func incrementSuccessfulRequests(website string) {
	mu.Lock()
	successfulRequests[website]++
	mu.Unlock()
}

func incrementFailedRequests(website string) {
	mu.Lock()
	failedRequests[website]++
	mu.Unlock()
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Fprintf(w, "total_requests %d\n", totalRequests)
	for site, count := range successfulRequests {
		fmt.Fprintf(w, "successful_requests{website=\"%s\"} %d\n", site, count)
	}
	for site, count := range failedRequests {
		fmt.Fprintf(w, "failed_requests{website=\"%s\"} %d\n", site, count)
	}
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
	rand.NewSource(time.Now().UnixNano())
	rand.Shuffle(len(websites), func(i, j int) {
		websites[i], websites[j] = websites[j], websites[i]
	})
	if count > len(websites) {
		count = len(websites)
	}
	return websites[:count]
}

func run(downloadFunc func(string) ([]string, error), sendFunc func(*http.Client, string)) {
	allWebsites, err := downloadFunc("https://analytics.usa.gov/data/live/sites.csv")
	if err != nil {
		log.Fatalf("Error downloading websites: %v", err)
	}

	resolver := &dnscache.Resolver{}
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			resolver.Refresh(true)
		}
	}()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			separator := strings.LastIndex(addr, ":")
			host := addr[:separator]
			ips, err := resolver.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no IPs found for host %s", host)
			}
			return net.Dial(network, ips[0]+addr[separator:])
		},
	}
	client := &http.Client{Transport: transport}

	quit := make(chan struct{})
	ticker := time.NewTicker(1 * time.Hour)

	wg := &sync.WaitGroup{}
	selectedWebsites := selectRandomWebsites(allWebsites, concurrency)

	for _, website := range selectedWebsites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			for {
				select {
				case <-quit:
					return
				default:
					sendFunc(client, site)
					time.Sleep(time.Duration(rand.Intn(delay-1)+1) * time.Second)
				}
			}
		}(website)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				close(quit)
				quit = make(chan struct{})
				wg.Wait()

				wg = &sync.WaitGroup{}
				selectedWebsites = selectRandomWebsites(allWebsites, concurrency)

				for _, website := range selectedWebsites {
					wg.Add(1)
					go func(site string) {
						defer wg.Done()
						for {
							select {
							case <-quit:
								return
							default:
								sendFunc(client, site)
								time.Sleep(time.Duration(rand.Intn(delay-1)+1) * time.Second)
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

func main() {
	var cmd = &cobra.Command{
		Use:   "foxtrot",
		Short: "A simple golang script that will help bump Firefox's overall numbers on US Gov websites",
		Run: func(cmd *cobra.Command, args []string) {
			go func() {
				http.Handle("/metrics", expvar.Handler())
				log.Fatal(http.ListenAndServe(":9120", nil))
			}()
			run(downloadWebsites, sendRequest)
		},
	}

	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "Number of goroutines and random websites to select")
	cmd.Flags().IntVarP(&delay, "delay", "d", 45, "Total time to sleep between requests in seconds")
	cmd.Flags().StringVarP(&userAgent, "user-agent", "u", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:123.0) Gecko/20100101 Firefox/123.0", "User-Agent for HTTP requests")

	if err := cmd.Execute(); err != nil {
		log.Fatalf("Command execution error: %v", err)
	}
}
