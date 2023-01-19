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
	URL         string
	Content     *[]byte
	ContentType string
}

func refresh(item *item) {
	fmt.Printf("Fetching %s\n", item.URL)

	resp, err := http.Get(item.URL)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", item.URL, err)
		return
	}

	item.ContentType = resp.Header.Get("Content-Type")
	item.Content = &b
}

var readmeStats item
var topLangs item
var streakStats item

var updates chan item

func worker(updates <-chan item) {
	fmt.Println("Register the worker")
	for item := range updates {
		fmt.Println("Worker processing job", item)
		refresh(&item)
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
	refresh(&readmeStats)
	refresh(&topLangs)
	refresh(&streakStats)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/stever/github-badge-cache", http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/readme-stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", readmeStats.ContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*readmeStats.Content)
		updates <- readmeStats
	})

	http.HandleFunc("/top-langs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", readmeStats.ContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*topLangs.Content)
		updates <- topLangs
	})

	http.HandleFunc("/streak-stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", readmeStats.ContentType)
		w.WriteHeader(http.StatusOK)
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
