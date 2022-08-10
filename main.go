package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

//var starting_page = "https://www.google.com/"
var starting_page = "https://blog.xkcd.com/"
var href_tag = "<a href="
var body_start_tag = "<body"
var body_end_tag = "</body>"

var searchIndex = make(map[string][]string)

func PrintIndex() {
	for keyword, urls := range searchIndex {

		fmt.Println("keyword")
		fmt.Println(keyword)
		fmt.Println("found here")
		for _, url := range urls {
			fmt.Println(url)
		}
	}
}

func SliceContainsElem(coll []string, val string) bool {
	for _, v := range coll {
		if v == val {
			return true
		}
	}
	return false
}

func addToIndex(keyword string, url string) {
	val, ok := searchIndex[keyword]
	if ok {
		if SliceContainsElem(val, url) == false {
			searchIndex[keyword] = append(val, url)
		}
	} else {
		urlList := make([]string, 0, 1)
		urlList = append(urlList, url)
		searchIndex[keyword] = urlList
	}

}

func removeTags(word string) (string, bool) {
	word = strings.Trim(word, "<p>")
	word = strings.Trim(word, "</p>")
	word = strings.Trim(word, "<em>")
	word = strings.Trim(word, "</em>")
	word = strings.Trim(word, ">")
	word = strings.Trim(word, "<")
	word = strings.Trim(word, "?")

	_, err := strconv.Atoi(word) //ignore numbers

	if strings.ContainsAny(word, ":<>,=\"-.+')(_[|;{}#") ||
		word == "" ||
		err == nil ||
		strings.HasPrefix(word, "class=") {
		return word, false

	}
	return word, true
}

func addPageToIndex(url string, pageContent string) {
	b := strings.Index(pageContent, body_start_tag)
	e := strings.Index(pageContent, body_end_tag)
	if b == -1 || e == -1 {
		return
	}
	pc := pageContent[b:e] //only search for words within body tags
	words := strings.Split(pc, " ")
	for _, word := range words {
		w, valid := removeTags(word)
		if valid {
			fmt.Println(w)
			addToIndex(word, url)
		}
	}
}

func searchKeyLookup(keyword string) {
	val, ok := searchIndex[keyword]
	if ok {
		for _, url := range val {
			fmt.Println(url)
		}
	} else {
		fmt.Printf("No results found for \"%s\"\n", keyword)
	}
}

func visitpage(page string) string {
	resp, err := http.Get(page)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func givenextlinkpos(s string, startpos int) int {
	idx := strings.Index(s[startpos:], href_tag)
	if idx > -1 {
		idx += startpos
	}
	return idx
}

func givenextquotepos(s string, slen int, si int) int {
	index := si

	for index < slen {
		if s[index] == 34 {
			return index
		}
		index++
	}
	return -1
}

func startCrawl(startPage string) {
	crawlIndex := 0
	crawlLimit := 2 //only crawl first 10 links
	crawled := make(map[string]bool)
	toCrawl := make([]string, 0, 100)

	toCrawl = append(toCrawl, startPage)

	for len(toCrawl) > 0 {
		page := toCrawl[len(toCrawl)-1]
		crawled[page] = true
		toCrawl = toCrawl[:len(toCrawl)-1]
		fmt.Println(page)

		pageBody := visitpage(page)
		if pageBody == "" {
			continue
		}

		addPageToIndex(page, pageBody)
		for _, cLink := range giveAllChidLinks(pageBody) {
			_, ok := crawled[cLink]
			if !ok {
				toCrawl = append(toCrawl, cLink)
			}
		}
		crawlIndex++
		if crawlIndex == crawlLimit {
			break
		}
	}
	fmt.Println("Index created")
}

func giveAllChidLinks(body string) []string {
	allLinks := make([]string, 0, 100)
	beglink := givenextlinkpos(body, 0)
	begquote := givenextquotepos(body, len(body), beglink+7)
	endquote := givenextquotepos(body, len(body), begquote+1)
	for begquote != -1 && endquote != -1 && beglink != -1 {
		url := body[begquote+1 : endquote]
		url = strings.TrimSpace(url)
		if url != "#" { //sometimes there are just #.
			allLinks = append(allLinks, url)
		}

		beglink = givenextlinkpos(body, endquote+1)
		begquote = givenextquotepos(body, len(body), beglink+7)
		endquote = givenextquotepos(body, len(body), begquote+1)
	}
	return allLinks
}

func main() {
	startCrawl(starting_page)
	for {
		var kw string
		fmt.Println("Enter a keyword to search...")
		fmt.Scanln(&kw)
		searchKeyLookup(kw)
	}
}
