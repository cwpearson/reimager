package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cwpearson/reddit-images/rate_limit"
	"github.com/cwpearson/reddit-images/reddit"
)

func main() {

	// Define and parse flags
	every := flag.Int64("every", 60*30, "Optional: Run every N seconds")
	outDir := flag.String("out-dir", "subreddits", "Optional: Output directory path")
	flag.Parse()

	subreddits := flag.Args()
	if len(subreddits) == 0 {
		fmt.Fprintf(os.Stderr, "Error: At least one string argument is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [--every N] [--out-dir path] string1 [string2 ...]\n", os.Args[0])
		os.Exit(1)
	}

	rl := rate_limit.NewRateLimit()

	for {
		for _, subreddit := range subreddits {
			r := reddit.NewReddit(rl, subreddit)
			r.Get(*outDir)
		}

		when := time.Now().Add(time.Second * time.Duration(*every))
		log.Println("sleep until", when)
		time.Sleep(time.Until(when))
	}

}
