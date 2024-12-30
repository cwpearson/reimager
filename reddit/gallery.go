package reddit

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cwpearson/reddit-images/rate_limit"
)

type GListing struct {
	Data GListingData `json:"data"`
}

type GListingData struct {
	Children []GListingDataChild `json:"children"`
}

type GListingDataChild struct {
	Data GListingDataChildData `json:"data"`
}

type GListingDataChildData struct {
	MediaMetadata map[string]Metadata `json:"media_metadata"`
}

type Metadata struct {
	Id       string `json:"id"`
	Mimetype string `json:"m"`
}

func GalleryImageMetadata(rl *rate_limit.RateLimit, url string) ([]Metadata, error) {
	jsonUrl := fmt.Sprintf("%s.json?raw_json=1", url)
	log.Printf("gallery url: %s -> %s", url, jsonUrl)

	content, err := rl.Get(jsonUrl, "")
	if err != nil {
		return nil, err
	}

	var data []GListing
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	res := []Metadata{}

	for _, val := range data[0].Data.Children[0].Data.MediaMetadata {
		res = append(res, val)
	}

	return res, nil
}
