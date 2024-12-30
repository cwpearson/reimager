package reddit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cwpearson/reddit-images/rate_limit"
	"github.com/gabriel-vasile/mimetype"
)

// Response represents the outer JSON structure
type Response struct {
	Data ListingData `json:"data"`
}

// ListingData represents the data field containing children
type ListingData struct {
	After    string  `json:"after"`
	Children []Child `json:"children"`
}

// Child represents each item in the children array
type Child struct {
	Kind string    `json:"kind"`
	Data ChildData `json:"data"`
}

// ChildData represents the nested data in each child
type ChildData struct {
	Title               string  `json:"title"`
	Author              string  `json:"author"`
	URLOverriddenByDest string  `json:"url_overridden_by_dest"`
	URL                 string  `json:"url"`
	Created             float64 `json:"created"`
	Id                  string  `json:"id"`
}

type Reddit struct {
	subreddit string
	retries   int
	rl        *rate_limit.RateLimit
}

func NewReddit(rl *rate_limit.RateLimit, subreddit string) *Reddit {
	return &Reddit{
		subreddit: subreddit,
		retries:   3,
		rl:        rl,
	}
}

// returns children, after, error
func (r *Reddit) Next(after string) ([]ChildData, string, error) {

	baseURL := fmt.Sprintf("https://reddit.com/r/%s/hot.json", r.subreddit)

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	// Create query parameters
	params := url.Values{}
	params.Add("raw_json", "1")
	params.Add("limit", "100")
	if after != "" {
		params.Add("after", after)
	}

	// Add the query parameters to the URL
	u.RawQuery = params.Encode()

	var body []byte
	for try := 0; try < r.retries; try++ {
		body, err = r.rl.Get(u.String(), "")
		if err != nil {
			fmt.Printf("Error getting subreddit: %v\n", err)
			body = nil
			time.Sleep(time.Second * time.Duration(5))
			continue
		}
		break
	}
	if body == nil {
		return nil, "", fmt.Errorf("retries exceeded")
	}
	response := Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, "", err
	}

	res := []ChildData{}
	for _, c := range response.Data.Children {
		if c.Kind == "t3" { // link
			res = append(res, c.Data)
		}
	}
	return res, response.Data.After, nil
}

func getImage(rl *rate_limit.RateLimit, url, outDir, stem string) error {
	contents, err := rl.Get(url, "image/*")
	if err != nil {
		return err
	}
	mtype := mimetype.Detect(contents)
	name := stem + mtype.Extension()
	outPath := filepath.Join(outDir, name)
	log.Println("write", outPath)
	return os.WriteFile(outPath, contents, 0644)
}

func (r *Reddit) Get(outDir string) {
	var children []ChildData
	var err error

	outDir = filepath.Join(outDir, r.subreddit)
	err = os.MkdirAll(outDir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Println("ERROR: couldn't create out directory", outDir)
		return
	}

	// load existing names
	existing := map[string]struct{}{}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		log.Println("ERROR: couldn't read directory", outDir)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		nameWithoutSuffix := strings.TrimSuffix(filename, filepath.Ext(filename))
		existing[nameWithoutSuffix] = struct{}{}
	}

	children, _, err = r.Next("")
	if err != nil {
		log.Println("ERROR: Next error:", err)
		return
	}

	for _, child := range children {
		log.Println("Title:", child.Title)
		shortTitle := child.Title
		if len(shortTitle) > 32 {
			shortTitle = shortTitle[0:32]
		}
		shortTitle = sanitizeFilename(shortTitle)

		if strings.Contains(child.URLOverriddenByDest, "www.reddit.com/gallery") {

			metas, err := GalleryImageMetadata(r.rl, child.URLOverriddenByDest)
			if err != nil {
				log.Println("ERROR: Gallery handling error:", err)
				continue
			}

			log.Println("Gallery metas:", metas)

			for mi, meta := range metas {
				parts := strings.Split(meta.Mimetype, "/")
				if len(parts) == 2 {
					imgUrl := fmt.Sprintf("https://i.redd.it/%s.%s", meta.Id, parts[1])

					stem := fmt.Sprintf("%d_%s_%s", int64(child.Created), shortTitle, meta.Id)

					if _, ok := existing[stem]; ok {
						log.Println(stem, "already downloaded")
						continue
					}

					err := getImage(r.rl, imgUrl, outDir, stem)
					if err != nil {
						log.Println("ERROR: getImage:", err)
						continue
					}
					existing[stem] = struct{}{}
				}
			}

			continue
		} else {

			stem := fmt.Sprintf("%d_%s_%s", int64(child.Created), shortTitle, child.Id)

			if _, ok := existing[stem]; ok {
				log.Println(stem, "already downloaded")
				continue
			}

			err := getImage(r.rl, child.URLOverriddenByDest, outDir, stem)
			if err != nil {
				log.Println("ERROR: getImage:", err)
				continue
			}
			existing[stem] = struct{}{}
		}
	}
}

func sanitizeFilename(input string) string {
	// Replace path separators and problematic characters
	replacer := strings.NewReplacer(
		"/", "",
		"\\", "",
		":", "",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
		",", "",
		";", "",
		"\x00", "", // null byte
		" ", "-", // replace spaces with hyphens
	)
	cleaned := replacer.Replace(input)

	// Remove non-ASCII characters
	var result strings.Builder
	for _, r := range cleaned {
		if r < 128 { // Keep only ASCII characters
			result.WriteRune(r)
		}
	}
	cleaned = result.String()

	// Trim spaces (though they should already be replaced with hyphens)
	cleaned = strings.TrimSpace(cleaned)

	// If the filename becomes empty after cleaning, provide a default
	if cleaned == "" {
		return "unnamed_file"
	}

	return cleaned
}
