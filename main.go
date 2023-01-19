package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type item struct {
	URL     string
	Content *[]byte
}

func get(url string) *[]byte {
	fmt.Printf("Fetching %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
		return nil
	}

	b, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", url, err)
		return nil
	}

	return &b
}

var readmeStats item
var topLangs item
var streakStats item

var updates chan item

func worker(updates <-chan item) {
	fmt.Println("Register the worker")
	for item := range updates {
		fmt.Println("Worker processing job", item)
		item.Content = get(item.URL)
	}
}

func main() {
	updates = make(chan item, 100)
	go worker(updates)

	// These are the badges that are going to be provided.
	readmeStats = item{URL: "https://github-readme-stats.stever.dev/api?username=stever&count_private=true&show_icons=true&theme=vision-friendly-dark&hide_title=true"}
	topLangs = item{URL: "https://github-readme-stats.stever.dev/api/top-langs/?username=stever&langs_count=10&layout=compact&theme=vision-friendly-dark&custom_title=Top%20Languages&hide=css,html,scss,Vim%20Script,PLpgSQL,NSIS,ANTLR,Dockerfile,LESS,Jupyter%20Notebook,CMake,QML,Batchfile,Makefile,Shell"}
	streakStats = item{URL: "https://github-readme-streak-stats.stever.dev?user=stever&theme=vision-friendly-dark&date_format=j%20M%5B%20Y%5D&mode=weekly"}

	// Pre-populate the cache.
	readmeStats.Content = get(readmeStats.URL)
	topLangs.Content = get(topLangs.URL)
	streakStats.Content = get(streakStats.URL)

	http.HandleFunc("/readme-stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write(*readmeStats.Content)
		updates <- readmeStats
	})

	http.HandleFunc("/top-langs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write(*topLangs.Content)
		updates <- topLangs
	})

	http.HandleFunc("/streak-stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write(*streakStats.Content)
		updates <- streakStats
	})

	port := 8080
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on port: %d\n", port)
	log.Fatal(s.ListenAndServe())
}
