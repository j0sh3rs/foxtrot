package foxtrot

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendRequest_Success(t *testing.T) {
	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("Expected User-Agent %v, got %v", userAgent, r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sendRequest(client, server.URL)
}

func TestSendRequest_ErrorCreatingRequest(t *testing.T) {
	client := &http.Client{}
	invalidURL := string([]byte{0x7f}) // Invalid URL to cause an error in http.NewRequest
	sendRequest(client, invalidURL)
}

func TestSendRequest_ErrorSendingRequest(t *testing.T) {
	client := &http.Client{}
	badServerURL := "http://bad.server.com"
	sendRequest(client, badServerURL)
}

func TestDownloadWebsites_Success(t *testing.T) {
	mockCSV := `example.com
example.org
example.net`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockCSV))
	}))
	defer server.Close()

	websites, err := downloadWebsites(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := []string{"https://example.com", "https://example.org", "https://example.net"}
	for i, site := range expected {
		if websites[i] != site {
			t.Errorf("Expected %v, got %v", site, websites[i])
		}
	}
}

func TestDownloadWebsites_Error(t *testing.T) {
	_, err := downloadWebsites("http://nonexistent.url")
	if err == nil {
		t.Fatalf("Expected an error, got none")
	}
}

func TestSelectRandomWebsites(t *testing.T) {
	websites := []string{"https://example1.com", "https://example2.com", "https://example3.com"}
	selected := selectRandomWebsites(websites, 2)

	if len(selected) != 2 {
		t.Errorf("Expected 2 websites, got %d", len(selected))
	}

	for _, site := range selected {
		if !contains(websites, site) {
			t.Errorf("Unexpected website selected: %v", site)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestRun(t *testing.T) {
	mockDownloadWebsites := func(url string) ([]string, error) {
		return []string{"https://example.com", "https://example.org"}, nil
	}

	mockSendRequest := func(client *http.Client, website string) {
		// Simulate a successful request
	}

	go func() {
		run(mockDownloadWebsites, mockSendRequest)
	}()

	time.Sleep(2 * time.Second)
}
