package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html"
)

func main() {
	// Track how long
	startTime := time.Now()
	var resultTotal uint64

	filepath := flag.String("path", "", "The filepath of the names")
	workers := flag.Int("workers", 2, "How many workers to run concurrently. (More workers are faster but more prone to rate limiting or bandwith issues)")
	sleep := flag.Int("sleep", 100, "Sleep duration between each workers task. (Millisecond)")
	// auth := flag.String("auth", "", "authenticity_token for post request to github")
	flag.Parse()

	var data []byte
	var err error
	if *filepath == "" {
		// Make sure something is being passed to Stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatal("Pass a filepath or data on stdin")
		}

		// If there is something let's read it all
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		data, err = ioutil.ReadFile(*filepath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Get auth token
	auth, cookie := getAuth()

	// Generate the word list from the filepath provided
	names, err := splitData(data)
	if err != nil {
		log.Fatal(err)
	}

	*workers = verifyWorkerCount(len(names), *workers)

	fmt.Printf("Starting github-usercheck. Parsing %d names with %d workers. \n", len(names), *workers)

	results := make(chan string)
	var wg sync.WaitGroup
	wg.Add(*workers)
	go func() {
		for r := range results {
			fmt.Println(r)

			// increment the result total
			atomic.AddUint64(&resultTotal, 1)
		}
		close(results)
	}()

	// Split names among workers
	for i := 0; i < *workers; i++ {
		go func(i int) {
			defer wg.Done()
			start, end := calculateLoad(len(names), *workers, i)
			for _, n := range names[start:end] {
				ok := available(n, auth, cookie)
				if ok {
					results <- n
					time.Sleep(time.Duration(*sleep) * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("Found %d results in %f seconds \n", resultTotal, time.Since(startTime).Seconds())
}

// verifyWorkerCount if there are fewer names than workers subtract workers
func verifyWorkerCount(totalLoad, workers int) int {
	ratio := float64(totalLoad) / float64(workers)
	if ratio == float64(int64(ratio)) {
		return workers
	}

	// Make sure there's more work than workers
	if ratio < 1 {
		for ratio < 1 {
			workers--
			ratio = float64(totalLoad) / float64(workers)
		}
		return workers
	}

	return workers
}

// calculateLoad divides the totalLoad among the number of workers based on turn
func calculateLoad(tasks, workers, turn int) (start, end int) {
	load := tasks / workers

	// Each turn the start and end index updated based on whose turn it is
	start = load * turn
	end = load * (turn + 1)
	if end < tasks {
		end = tasks
	}

	return
}

// splitData takes a slice of bytes and splits on \n into a slice of strings
func splitData(data []byte) ([]string, error) {
	names := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		// Don't append empty lintes
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}

	return names, nil
}

// getAuth makes a requst to github, parses the html response for the authenticity_token
// and returns any cookies
func getAuth() (token string, cookies []*http.Cookie) {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	res, err := c.Get("https://github.com/session")
	if err != nil {
		return "", nil
	}
	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		return "", nil
	}

	// Look through html to find authenticity_token
	var auth string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Data == "input" {
			for _, a := range n.Attr {
				if a.Key == "name" {
					if a.Val == "authenticity_token" {
						// Now find the value of the name
						for _, a := range n.Attr {
							if a.Key == "value" {
								auth = a.Val
							}
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// fmt.Println(res.Header)
	cookies = res.Cookies()

	return auth, cookies
}

// available check github.com and verifies if the name arg is available
// if there is an auth strig use signup_check form
func available(name, auth string, cookies []*http.Cookie) bool {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	apiURL := "https://github.com"
	resource := "/signup_check/username"
	data := url.Values{}
	data.Add("value", name)
	data.Add("authenticity_token", auth)

	u, err := url.ParseRequestURI(apiURL)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Referer", "https://github.com/join?source=header-home")

	for _, c := range cookies {
		req.AddCookie(c)
	}

	res, err := c.Do(req)
	if err != nil {
		return false
	}

	if res.StatusCode != http.StatusOK {
		return false
	}

	return true
}
