package main

import (
	"strings"
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New([]byte(strings.ToLower(s.CompleteWorks)))
	return nil
}

func (s *Searcher) Search(query string) []string {
	idxs := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, SemiMeaningfulSlice(s.CompleteWorks, idx))
	}
	return results
}

func SemiMeaningfulSlice(s string, from int) string {
	const period byte = '.'
	const semicolon byte = ';'
	const openingBracket byte = '['
	const closingBracket byte = ']'
	prevIdx := 0;
	for i := from - 100; i >= 0; i-- {
		if (s[i] == period) || (s[i] == semicolon) {
			prevIdx = i+1;
			break;
		}
	} 

	nextIdx := 0
	for i := from + 100; i < len(s); i++ {
		if s[prevIdx] == openingBracket && s[i+1] == closingBracket {
			nextIdx = i+1;
			break
		} else if (s[i] == period) || (s[i] == semicolon) {
			nextIdx = i+1;
			break
		}
	}

	println("from: " + strconv.Itoa(from))
	println("prevIdx: " + strconv.Itoa(prevIdx))
	println("nextIdx: " + strconv.Itoa(nextIdx))

	return s[prevIdx:nextIdx]
}
