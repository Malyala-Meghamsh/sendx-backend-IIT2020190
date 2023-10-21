package routes

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	// "golang.org/x/text/date"
)

var paidJobsMutex sync.Mutex

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

type job struct {
	url           string
	accessed_time time.Time
	isUserPaid    bool
	resWriter     *http.ResponseWriter
}

var cache = NewURLCache()

func main() {}

var paidWorkers = 5
var unpaidWorkers = 2
var jobsPaid = make(chan job, paidWorkers)

// var jobsUnpaid = make(chan job, unpaidWorkers)

func Main() {
	// cache.Add("https://www.youtube.com/watch?v=91EzD9VgwGk", time.Now())
	paidWorkers = 5
	unpaidWorkers = 2
	for i := 1; i <= paidWorkers; i++ {
		// wg.Add(1)
		go paidURLFetcher(jobsPaid)
	}
	for i := 1; i <= unpaidWorkers; i++ {
		// wg.Add(1)
		// go worker(i, requests, results, &wg)
	}
}

func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	// fmt.Println(query["isPaidCustomer"][0])
	isPaidCustomer := query["isPaidCustomer"][0]
	// job <- 5;
	// <-job
	if isPaidCustomer == "1" {
		w.Header().Set("Content-Type", "text/html")
		go servePaidUser(job{
			url:           query["url"][0],
			accessed_time: time.Now(),
			isUserPaid:    true,
			resWriter:     &w}) //ptr also try
	} else {
		// go serveUnpaidUser(job{
		// 	url:           query["url"][0],
		// 	accessed_time: time.Now(),
		// 	isUserPaid:    false})
	}
	// res, _ := http.Get(query["url"][0])
	// body, _ := io.ReadAll(res.Body)
	// fmt.Fprint(w, string(body))
}

func servePaidUser(j job) {
	// timeDiff := endTime.Sub(startTime)
	temp := j.url
	if cache.Contains(j.url) && time.Now().Sub(cache.url[temp]).Minutes() <= 60 {
		// fmt.Println(cache.url[temp])
		fmt.Println("Yes")
	} else {
		jobsPaid <- j
		// cache.Add(j.url, time.Now())
		fmt.Println("Added")
	}
}

func paidURLFetcher(jobs chan job) {
	for {
		paidJobsMutex.Lock()
		job := <-jobs
		paidJobsMutex.Unlock()
		fmt.Println("Fetching Real Time")
		res, err := http.Get(job.url)
		if err != nil {
			fmt.Println("Error fetching URL:", err)
			continue
		}
		fmt.Println("Upto now no error")
		defer res.Body.Close()
		response, err := io.ReadAll(res.Body)
		fmt.Fprintf((*job.resWriter), string(response))
		// job.resWriter.WriteHeader(http.StatusOK)
		// _, err = io.Copy(*job.resWriter, "Success")
		// _, err = fmt.Fprint(*job.resWriter, "Success ra")
		// (*job.resWriter).Write([]byte("success ra"))
		// *job.resWriter.Header.Set("Content-Type", "application/json")
		// .Header().Set("Content-Type", "application/json")
		// json.NewEncoder(*job.resWriter).Encode("Hi")
		// _, err = io.Copy(*job.resWriter, res.Body)
		// if err != nil {
		// 	fmt.Println("Error copying response body:", err)
		// }
		// _, err1 := io.ReadAll(res.Body)
		// defer res.Body.Close()
		// if err != nil {
		// 	fmt.Println("Error reading response body:", err)
		// 	continue
		// }
		// // // fmt.Println("Upto now no error part 2")
		// // job.resWriter.Header().Set("Content-Type", "text/html")
		// // // fmt.Fprint(job.resWriter, string(body))
		// // // fmt.Println("fine")
		// // fmt.Println(res.StatusCode)
		// // job.resWriter.WriteHeader(http.StatusOK) // Set HTTP status code if necessary
		// // // fmt.Println(body)
		// // _, err = job.resWriter.Write([]byte("res.StatusCode"))
		// // if err != nil {
		// // 	fmt.Println("Error writing response body:", err)
		// // }
		fmt.Println("No Error until this step")
		// Close the response writer
		// if closer, ok := job.resWriter.(io.Closer); ok {
		// 	closer.Close()
		// }
	}
}
