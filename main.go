package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

// TODO: should I use a Markdown parser?
func getLinks(cont []byte) []string {
	reLink := regexp.MustCompile(`\[.+\]\((\S+\(?\S+\)?)\)`)
	matches := reLink.FindAllStringSubmatch(string(cont), -1)
	var links []string
	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			// TODO: have a general function to clean up hyperlink?
			if strings.HasPrefix(link, "<") && strings.HasSuffix(link, ">") {
				link = strings.TrimLeft(link, "<")
				link = strings.TrimRight(link, ">")
			}
			links = append(links, link)
			// TODO: create a filter to skip all the "known" exceptions
		}
	}
	return links
}

// Use HTTP inspired status code return for relative path as well
func checkLink(link string, filePath string) int {
	if strings.HasPrefix(link, "http") {
		res, err := http.Get(link)
		if err != nil {
			return 500
		}
		return res.StatusCode
	}
	linkAbsPath := path.Join(path.Dir(filePath), link)
	if _, err := os.Stat(linkAbsPath); err != nil {
		// Keep it simple: assume file is missing
		return 404
	}
	return 200
}

func checkFilePath(filePath string) {
	cont, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading %s: %v", filePath, err)
	}

	links := getLinks(cont)
	ch := make(chan map[string]int)
	for _, link := range links {
		go func(text string) {
			ch <- map[string]int{link: checkLink(link, filePath)}
		}(link)
	}

	for range links {
		for link, status := range <-ch {
			if status != 200 {
				log.Printf("(%s): %q -> status %d\n", filePath, link, status)
			}
		}
	}
}

func main() {
	input, ok := os.LookupEnv("INPUT_FILEPATH")
	if !ok {
		return
	}
	filePaths := strings.Split(input, " ")
	done := make(chan bool)
	for _, filePath := range filePaths {
		go func(filepath string) {
			checkFilePath(filepath)
			done <- true
		}(filePath)
	}
	for range filePaths {
		<-done
	}
}
