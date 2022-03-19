package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const username = "lemon-mint"

func FetchStars(username string) ([]Star, error) {
	var stars []Star
	var i int
	for {
		i++
		url := fmt.Sprintf("https://api.github.com/users/%s/starred?per_page=100&page=%d", username, i)
		fmt.Printf("Fetching %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("Status code %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		var page []Star
		err = json.NewDecoder(resp.Body).Decode(&page)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		stars = append(stars, page...)
	}
	return stars, nil
}

func main() {
	stars, err := FetchStars(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Fetched %d stars\n", len(stars))
	for _, star := range stars {
		fmt.Printf("%+v\n", star.Name)
	}
}
