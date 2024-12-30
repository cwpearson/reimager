package rate_limit

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type RateLimit struct {
	client    *http.Client
	used      int
	remaining int
	reset     time.Time
}

func getFutureTime(seconds int) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

func sleepUntil(when time.Time) {
	now := time.Now()
	if now.After(when) {
		return
	}
	time.Sleep(when.Sub(now))
}

func NewRateLimit() *RateLimit {
	return &RateLimit{
		client:    &http.Client{},
		remaining: 100,
		reset:     time.Now(),
	}
}

func (rl *RateLimit) UpdateUsed(used string) {
	val, err := strconv.ParseFloat(used, 64)
	if err == nil {
		rl.used = int(val)
		log.Println("used ->", rl.used)
	}
}

func (rl *RateLimit) UpdateRemaining(used string) {
	val, err := strconv.ParseFloat(used, 64)
	if err == nil {
		rl.remaining = int(val)
		log.Println("remaining ->", rl.remaining)
	}
}

func (rl *RateLimit) UpdateReset(used string) {
	val, err := strconv.ParseFloat(used, 64)
	if err == nil {
		maybe := getFutureTime(int(val))
		if rl.reset.Before(maybe) {
			rl.reset = maybe
			log.Println("reset ->", rl.reset)
		}
	}
}

// Get makes an HTTP GET request to the specified URL while respecting rate limits
func (rl *RateLimit) Get(url string, accept string) ([]byte, error) {
	if rl.remaining <= 0 {
		log.Println("no requests remaining, sleep until", rl.reset)
		sleepUntil(rl.reset)
	}

	// Create new request
	log.Println("GET", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Add any required headers
	req.Header.Add("User-Agent", "linux:reddit-images:0.1")
	if accept != "" {
		req.Header.Add("Accept", accept)
	}

	// Make the request
	resp, err := rl.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// parse rate limit
	used := resp.Header.Get("X-Ratelimit-Used")
	remaining := resp.Header.Get("X-Ratelimit-Remaining")
	reset := resp.Header.Get("X-Ratelimit-Reset")
	// fmt.Printf("User Limit: %s\n", used)
	// fmt.Printf("Remaining: %s\n", remaining)
	// fmt.Printf("Reset: %s\n", reset)
	rl.UpdateUsed(used)
	rl.UpdateRemaining(remaining)
	rl.UpdateReset(reset)

	if resp.StatusCode == http.StatusTooManyRequests {
		rl.remaining = 0

		// if the reset time is before now, just pick a while in the future for the next retry
		if rl.reset.Before(time.Now()) {
			rl.UpdateReset("450")
		}
		return nil, fmt.Errorf("request failed with 429")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
