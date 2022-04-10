package main

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

/*
This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <https://unlicense.org>
*/

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

var LinkCounter map[string]int = make(map[string]int)

var HTMLTagRegex = regexp.MustCompile(`<[^>]*>`)
var duplicateHyphenRegex = regexp.MustCompile(`-+`)

func GetAnchorLink(title string) string {
	// Remove any leading or trailing whitespace
	title = strings.TrimSpace(title)

	// Convert to lowercase
	title = strings.ToLower(title)

	// Remove any non-word characters
	title = strings.Trim(title, "!@#$%^&*()-_+={}[]|\\:;'<>?,./\"")
	// (remove HTML tags)
	title = HTMLTagRegex.ReplaceAllString(title, "")

	// Replace spaces with hyphens
	title = strings.Replace(title, " ", "-", -1)

	// Remove any duplicate hyphens

	// Duplicate hyphens REGEX
	title = duplicateHyphenRegex.ReplaceAllString(title, "-")

	// Remove any leading or trailing hyphens
	title = strings.Trim(title, "-")

	// Add the counter
	LinkCounter[title]++

	if LinkCounter[title] > 1 {
		title = fmt.Sprintf("#%s-%d", title, LinkCounter[title]-1)
		return title
	}

	return "#" + title
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
		//h := sha256.Sum256([]byte(v.Key))
		//h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))
		//sb.WriteString(fmt.Sprintf("* [%s](#v-%s)\n", v.Key, h32))

		sb.WriteString(fmt.Sprintf("* [%s](#%s)\n", v.Key, GetAnchorLink(v.Key)))
	}
	sb.WriteString("\n")

	for _, v := range languages {
		h := sha256.Sum256([]byte(v.Key))
		h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))

		sb.WriteString("<a name=\"v-" + h32 + "\"></a>\n")
		sb.WriteString(fmt.Sprintf("# %s\n\n", v.Key))

		for _, star := range v.Value {
			//h := sha256.Sum256([]byte(star.FullName))
			//h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))
			//sb.WriteString(fmt.Sprintf("* [%s](#repo-%s)\n", star.FullName, h32))
			sb.WriteString(fmt.Sprintf("* [%s](%s)\n", star.FullName, GetAnchorLink(star.FullName)))
		}
		sb.WriteString("\n")

		for _, star := range v.Value {
			h := sha256.Sum256([]byte(star.FullName))
			h32 := strings.ToLower(base32.StdEncoding.EncodeToString(h[:15]))
			sb.WriteString(fmt.Sprintf("<a name=\"repo-%s\"></a>\n", h32))
			//sb.WriteString(fmt.Sprintf("## [%s](%s)\n\n", star.FullName, star.HTMLURL))
			sb.WriteString(fmt.Sprintf("## %s\n\n", star.FullName))
			sb.WriteString(fmt.Sprintf("Repository: [%s](%s)\n", star.FullName, star.HTMLURL))
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
