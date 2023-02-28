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
	Header  map[string][]string
}

func refresh(item *item) {
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

	item.Content = &b
	item.Header = resp.Header
}

var readmeStats item
var topLangs item
var streakStats item
var emailShield item
var linkedinShield item
var mastoShield item
var ghRepoShield item
var ghGistsShield item
var ghStarsShield item
var ghPackagesShield item

var updates chan item

func worker(updates <-chan item) {
	for item := range updates {
		refresh(&item)
	}
}

func setHeaders(w http.ResponseWriter, header map[string][]string) {
	for headerName, valueArray := range header {
		for _, value := range valueArray {
			w.Header().Set(headerName, value)
		}
	}
}

func main() {
	updates = make(chan item, 100)
	go worker(updates)

	// These are the badges that are going to be provided.
	readmeStats = item{URL: "https://github-readme-stats.vercel.app/api?username=stever&count_private=true&show_icons=true&theme=vision-friendly-dark&hide_title=true"}
	topLangs = item{URL: "https://github-readme-stats.vercel.app/api/top-langs/?username=stever&langs_count=10&layout=compact&theme=vision-friendly-dark&custom_title=Top%20Languages&hide=css,html,scss,Vim%20Script,PLpgSQL,NSIS,ANTLR,Dockerfile,LESS,Jupyter%20Notebook,CMake,QML,Batchfile,Makefile,Shell"}
	streakStats = item{URL: "https://streak-stats.demolab.com/?user=stever&theme=vision-friendly-dark&date_format=j%20M%5B%20Y%5D&mode=weekly"}
	emailShield = item{URL: "https://img.shields.io/badge/-stever%40hey.com-5522fa?style=flat&label=&labelColor=white&logo=Hey&logoColor=5522fa"}
	linkedinShield = item{URL: "https://img.shields.io/badge/-csteve-2266c2?style=flat&logo=Linkedin&logoColor=white"}
	mastoShield = item{URL: "https://img.shields.io/badge/-%40stever%40hachyderm.io-5538c7?style=flat&label=&labelColor=white&logo=Mastodon&logoColor=5538c7"}
	ghRepoShield = item{URL: "https://img.shields.io/badge/-Repositories-silver?style=flat&label=&labelColor=black&logo=GitHub&logoColor=silver"}
	ghGistsShield = item{URL: "https://img.shields.io/badge/-Gists-silver?style=flat&label=&labelColor=black&logo=GitHub&logoColor=silver"}
	ghStarsShield = item{URL: "https://img.shields.io/badge/-Stars-silver?style=flat&label=&labelColor=black&logo=GitHub&logoColor=silver"}
	ghPackagesShield = item{URL: "https://img.shields.io/badge/-Packages-silver?style=flat&label=&labelColor=black&logo=GitHub&logoColor=silver"}

	// Pre-populate the cache.
	refresh(&readmeStats)
	refresh(&topLangs)
	refresh(&streakStats)
	refresh(&emailShield)
	refresh(&linkedinShield)
	refresh(&mastoShield)
	refresh(&ghRepoShield)
	refresh(&ghGistsShield)
	refresh(&ghStarsShield)
	refresh(&ghPackagesShield)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/stever/github-badge-cache", http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/readme-stats", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, readmeStats.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*readmeStats.Content)
		updates <- readmeStats
	})

	http.HandleFunc("/top-langs", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, topLangs.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*topLangs.Content)
		updates <- topLangs
	})

	http.HandleFunc("/streak-stats", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, streakStats.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*streakStats.Content)
		updates <- streakStats
	})

	http.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, emailShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*emailShield.Content)
		updates <- emailShield
	})

	http.HandleFunc("/linkedin", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, linkedinShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*linkedinShield.Content)
		updates <- linkedinShield
	})

	http.HandleFunc("/mastodon", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, mastoShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*mastoShield.Content)
		updates <- mastoShield
	})

	http.HandleFunc("/gh-repositories", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, ghRepoShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*ghRepoShield.Content)
		updates <- ghRepoShield
	})

	http.HandleFunc("/gh-gists", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, ghGistsShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*ghGistsShield.Content)
		updates <- ghGistsShield
	})

	http.HandleFunc("/gh-stars", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, ghStarsShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*ghStarsShield.Content)
		updates <- ghStarsShield
	})

	http.HandleFunc("/gh-packages", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, ghPackagesShield.Header)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(*ghPackagesShield.Content)
		updates <- ghPackagesShield
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
