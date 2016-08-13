package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// Track how long
	startTime := time.Now()
	var resultTotal uint64

	filepath := flag.String("path", "", "The filepath of the names")
	workers := flag.Int("workers", 2, "How many workers to run in parallel. (More scrapers are faster, but more prone to rate limiting or bandwith issues)")
	sleep := flag.Int("sleep", 100, "Sleep duration between each workers task. (Millisecond)")
	auth := flag.String("auth", "", "authenticity_token for post request to github")
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
				ok := available(n, *auth)
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

// If there are fewer names than workers subtract workers and if the totalLoad
// can't be evenly divided among worker subtract workers
func verifyWorkerCount(totalLoad, workers int) int {
	ratio := float64(totalLoad) / float64(workers)
	if ratio == 1 {
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

	// If the work can't be evenly divided among the workers in whole numbers
	// then reduce the number of workers
	if ratio != float64(int64(ratio)) {
		for ratio != float64(int64(ratio)) {
			workers--
			ratio = float64(totalLoad) / float64(workers)
		}

		return workers
	}

	return workers
}

// calculateLoad divides the totalLoad among the number of workers based on turn
func calculateLoad(totalLoad, workers, turn int) (start, end int) {
	load := totalLoad / workers

	// Each turn the start and end index updated based on whose turn it is
	start = load * turn
	end = load * (turn + 1)

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

// available check github.com and verifies if the name arg is available
// if there is an auth strig use signup_check form
func available(name string, auth string) bool {
	c := &http.Client{
		Timeout: time.Second * 10,
	}

	var res *http.Response
	var err error
	if auth == "" {
		res, err = c.Get(fmt.Sprintf("https://github.com/%s", name))
		if err != nil {
			return false
		}

		if res.StatusCode == http.StatusOK {
			return false
		}
	} else {
		v := url.Values{}
		v.Set("value", name)
		res, err = c.PostForm("https://github.com/signup_check/username", v)
		if err != nil {
			return false
		}

		if res.StatusCode != http.StatusOK {
			return false
		}
	}

	return true
}
