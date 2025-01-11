package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func loadDataFromCSV(filePath string, trie *Trie) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(record) < 2 {
			continue
		}

		word := strings.TrimSpace(record[0])
		meanings := []string{}
		antonyms := []string{}

		re := regexp.MustCompile("\\d+\\s+([^\\d]+)")
		matches := re.FindAllStringSubmatch(record[1], -1)

		if len(matches) == 0 {
			// Fallback logic if no matches
			meaningsRaw := strings.Split(record[1], "&")
			meanings = strings.Split(meaningsRaw[0], ",")
			if len(meaningsRaw) > 1 {
				antonyms = strings.Split(meaningsRaw[1], ",")
			}
		} else {
			for _, match := range matches {
				meaningsRaw := strings.Split(match[1], "&")
				for _, meaning := range strings.Split(meaningsRaw[0], ",") {
					meanings = append(meanings, meaning)
				}
				if len(meaningsRaw) > 1 {
					for _, antonym := range strings.Split(meaningsRaw[1], ",") {
						antonyms = append(antonyms, antonym)
					}
				}
			}
		}

		metadata := map[string][]string{
			"meanings": meanings,
			"antonyms": antonyms,
		}

		trie.Insert(word, metadata)
	}

	return nil
}

func loadData() {
	if _, err := os.Stat("trie_data.json"); err == nil {
		data, readErr := os.ReadFile("trie_data.json")
		if readErr != nil {
			panic("Failed to read trie_data.json: " + readErr.Error())
		}
		trie, deserErr := DeserializeTrie(data)
		if deserErr != nil {
			panic("Failed to deserialize trie_data.json: " + deserErr.Error())
		}
		globalTrie = trie
	} else {
		csvErr := loadDataFromCSV("words.csv", globalTrie)
		if csvErr != nil && !errors.Is(csvErr, os.ErrNotExist) {
			fmt.Println("Warning: Failed to load CSV (words.csv) -", csvErr)
		}
		data, serErr := globalTrie.Serialize()
		if serErr == nil {
			_ = os.WriteFile("trie_data.json", data, 0644)
		}
	}
}
