package main

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

const username = "lemon-mint"

func FetchStars(username string) ([]Star, error) {
	var stars []Star
	var i int
	for {
		i++
		url := fmt.Sprintf("https://api.github.com/users/%s/starred?per_page=100&page=%d", username, i)
		fmt.Fprintf(os.Stderr, "Fetching %s\n", url)
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

type starSorter []Star

func (s starSorter) Len() int {
	return len(s)
}

func (s starSorter) Less(i, j int) bool {
	return s[i].FullName < s[j].FullName
}

func (s starSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type KV[T any] struct {
	Key   string
	Value T
}

type KVSorter[T any] []KV[T]

func (s KVSorter[T]) Len() int {
	return len(s)
}

func (s KVSorter[T]) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func (s KVSorter[T]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func BuildMarkdown(stars []Star) string {
	var sortedByLanguage map[string][]Star = make(map[string][]Star)
	var sb strings.Builder

	for _, star := range stars {
		if star.Language == "" {
			star.Language = "Unknown"
		}
		sortedByLanguage[star.Language] = append(sortedByLanguage[star.Language], star)
	}
	var languages []KV[[]Star]

	for k := range sortedByLanguage {
		sort.Sort(starSorter(sortedByLanguage[k]))
		// println(k)
		// for _, star := range sortedByLanguage[k] {
		// 	println(star.FullName)
		// }
		if k != "Unknown" {
			languages = append(languages, KV[[]Star]{k, sortedByLanguage[k]})
		}
	}
	sort.Sort(KVSorter[[]Star](languages))
	languages = append(languages, KV[[]Star]{"Unknown", sortedByLanguage["Unknown"]})

	sb.WriteString("# Starred Repositories\n\n")
	sb.WriteString(fmt.Sprintf("This is a list of repositories starred by [%s](https://github.com/%s).\n\n", username, username))
	sb.WriteString("# Table of Contents\n\n")
	for _, v := range languages {
		// Table of contents
		h := sha256.Sum256([]byte(v.Key))
		h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))

		sb.WriteString(fmt.Sprintf("* [%s](#v-%s)\n", v.Key, h32))
	}
	sb.WriteString("\n")

	for _, v := range languages {
		h := sha256.Sum256([]byte(v.Key))
		h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))

		sb.WriteString("<a name=\"v-" + h32 + "\"></a>\n")
		sb.WriteString(fmt.Sprintf("# %s\n\n", v.Key))
		for _, star := range v.Value {
			sb.WriteString(fmt.Sprintf("## [%s](%s)\n\n", star.FullName, star.HTMLURL))
			sb.WriteString(fmt.Sprintf("Author: [%s](%s)\n\n", star.Owner.Login, star.Owner.HTMLURL))
			sb.WriteString(fmt.Sprintf("Stars: %d\n\n", star.StargazersCount))
			sb.WriteString(star.Description)
			sb.WriteString("\n\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func main() {
	stars, err := FetchStars(username)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(os.Stderr, "Fetched %d stars\n", len(stars))
	markdown := BuildMarkdown(stars)
	fmt.Println(markdown)
}
