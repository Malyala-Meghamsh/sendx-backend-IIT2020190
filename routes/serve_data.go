package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	// "golang.org/x/text/date"
)

type URLCache struct {
	mu  sync.RWMutex
	url map[string]time.Time
}

func NewURLCache() *URLCache {
	return &URLCache{
		url: make(map[string]time.Time),
	}
}

func (c *URLCache) Add(new_url string, t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.url[new_url] = t
}

func (c *URLCache) Contains(url string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.url[url]
	return exists
}

func main() {
	// for i := 1; i <= paidWorkers; i++ {
	// 	// wg.Add(1)
	// 	go worker()
	// }
	// for i := 1; i <= unpaidWorkers; i++ {
	// 	// wg.Add(1)
	// 	// go worker(i, requests, results, &wg)
	// }
}

var cache = NewURLCache()

func Main() {
	cache.Add("https://www.youtube.com/watch?v=91EzD9VgwGk", time.Now())
	// cache["https://www.youtube.com/watch?v=91EzD9VgwGk"] = true
	for i := 1; i <= paidWorkers; i++ {
		// wg.Add(1)
		// go paidURLFetcher(jobsPaid)
	}
	for i := 1; i <= unpaidWorkers; i++ {
		// wg.Add(1)
		// go worker(i, requests, results, &wg)
	}
}

// type jobPaid struct {
// 	url           string
// 	accessed_time time.Time
// }

type job struct {
	url           string
	accessed_time time.Time
	isUserPaid    bool
	resWriter     http.ResponseWriter
}

var paidWorkers = 5
var unpaidWorkers = 2
var jobsPaid = make(chan job, paidWorkers)
var jobsUnpaid = make(chan job, unpaidWorkers)

func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// fmt.Println(query["isPaidCustomer"][0])
	isPaidCustomer := query["isPaidCustomer"][0]
	// job <- 5;
	// <-job
	if isPaidCustomer == "1" {
		go servePaidUser(job{
			url:           query["url"][0],
			accessed_time: time.Now(),
			isUserPaid:    true,
			resWriter:     w}) //ptr also try
	} else {
		// go serveUnpaidUser(job{
		// 	url:           query["url"][0],
		// 	accessed_time: time.Now(),
		// 	isUserPaid:    false})
	}
	// res, _ := http.Get(query["url"][0])
	// body, _ := io.ReadAll(res.Body)
	// w.Header().Set("Content-Type", "text/html")
	// fmt.Fprint(w, string(body))
}

func servePaidUser(j job) {
	// timeDiff := endTime.Sub(startTime)
	temp := j.url
	if cache.Contains(j.url) && time.Now().Sub(cache.url[temp]).Minutes() <= 60 {
		// fmt.Println(cache.url[temp])
		fmt.Println("Yes")

	} else {
		cache.Add(j.url, time.Now())
		fmt.Println("Added")
	}
}
