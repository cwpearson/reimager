package main

import (
	"log"
	"time"

	"github.com/cwpearson/reddit-images/rate_limit"
	"github.com/cwpearson/reddit-images/reddit"
)

func main() {

	subreddits := []string{
		"pics",
		"oldschoolcool",
		"thewaywewere",
		"MilitaryPorn",
		"EarthPorn",
	}

	rl := rate_limit.NewRateLimit()

	for {
		for _, subreddit := range subreddits {
			r := reddit.NewReddit(rl, subreddit)
			r.Get()
		}

		when := time.Now().Add(time.Minute * time.Duration(30))
		log.Println("sleep until", when)
		time.Sleep(time.Until(when))
	}

}
