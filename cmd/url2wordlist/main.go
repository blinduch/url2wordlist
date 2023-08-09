package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func fetchWordsFromURL(urlStr string, wg *sync.WaitGroup, resultChan chan<- []string) {
	defer wg.Done()

	// Create a context with a timeout of 5 seconds for the URL operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		resultChan <- nil
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		resultChan <- nil
		return
	}
	defer resp.Body.Close()

	var words []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		resultChan <- nil
		return
	}

	validWordRegex := regexp.MustCompile(`[\w-]+`)
	filteredWords := make([]string, 0)

	for _, line := range words {
		matches := validWordRegex.FindAllString(line, -1)
		filteredWords = append(filteredWords, matches...)
	}

	resultChan <- filteredWords
}

func main() {
	urls := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	var wg sync.WaitGroup
	resultChan := make(chan []string)

	for _, urlStr := range urls {
		wg.Add(1)

		go fetchWordsFromURL(urlStr, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allWords []string
	for words := range resultChan {
		if words != nil {
			allWords = append(allWords, words...)
		}
	}

	wordCount := make(map[string]int)
	for _, word := range allWords {
		word = strings.ToLower(word)
		wordCount[word]++
	}

	var sortedWords []string
	for word := range wordCount {
		sortedWords = append(sortedWords, word)
	}

	// Sort words by occurrence
	sortWordsByOccurrence(sortedWords, wordCount)

	// Print the sorted words
	for _, word := range sortedWords {
		fmt.Println(word)
	}
}

func sortWordsByOccurrence(words []string, wordCount map[string]int) {
	sort.Slice(words, func(i, j int) bool {
		return wordCount[words[i]] > wordCount[words[j]]
	})
}

