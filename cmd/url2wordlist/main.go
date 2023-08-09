package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

func fetchWordsFromURL(url string, wg *sync.WaitGroup, resultChan chan<- []string) {
	defer wg.Done()

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching URL %s: %s\n", url, err)
		resultChan <- nil
		return
	}
	defer response.Body.Close()

	var words []string
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Use a regular expression to find valid words (alphanumeric, underscore, hyphen)
		validWordRegex := regexp.MustCompile(`[\w-]+`)
		matches := validWordRegex.FindAllString(line, -1)
		words = append(words, matches...)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading response from URL %s: %s\n", url, err)
		resultChan <- nil
		return
	}

	resultChan <- words
}

func mergeWords(wordLists [][]string) []string {
	merged := make([]string, 0)
	for _, wordList := range wordLists {
		merged = append(merged, wordList...)
	}
	return merged
}

func countWords(words []string) map[string]int {
	wordCount := make(map[string]int)
	for _, word := range words {
		word = strings.ToLower(word)
		wordCount[word]++
	}
	return wordCount
}

func sortByOccurrence(wordCount map[string]int) []string {
	type wordFrequency struct {
		word      string
		occurrence int
	}

	var wordFreqList []wordFrequency
	for word, count := range wordCount {
		wordFreqList = append(wordFreqList, wordFrequency{word, count})
	}

	sort.Slice(wordFreqList, func(i, j int) bool {
		return wordFreqList[i].occurrence > wordFreqList[j].occurrence
	})

	var sortedWords []string
	for _, wf := range wordFreqList {
		sortedWords = append(sortedWords, wf.word)
	}

	return sortedWords
}

func main() {
	urls := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	var wg sync.WaitGroup
	resultChan := make(chan []string)

	for _, url := range urls {
		wg.Add(1)
		go fetchWordsFromURL(url, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var allWords [][]string
	for words := range resultChan {
		if words != nil {
			allWords = append(allWords, words)
		}
	}

	mergedWords := mergeWords(allWords)
	wordCount := countWords(mergedWords)
	sortedWords := sortByOccurrence(wordCount)

	for _, word := range sortedWords {
		fmt.Println(word)
	}
}

