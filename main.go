package main

import (
	"fmt"
	"math"
	"net/http"
	"sendx/fetch_content"
	"strconv"
	"sync"
	"time"
)

type Job struct { // Convert the URL to a job
	URL            string
	ResultChannel  chan string // Storing the result in this channel
	isPaidCustomer bool
	retries        int
}

var (
	MaxPaidWorkers    = 5 // Workers for crawling for paid Users
	MaxNonPaidWorkers = 2 // Workers for crawling for un-paid Users

	maxRequestsPerHour = math.MaxInt64 // Setting max requests initially as Maximum value

	reqMutex sync.Mutex     // Mutex to control the incoming requests if ourworkers are shutdown
	wg       sync.WaitGroup // To check if all our workers are shutdown

	// Seperate JobQueue for paid and nonpaid workers
	paidJobQueue    chan Job
	nonPaidJobQueue chan Job

	shutdown chan struct{} // To give signal for workers to stop
)

// Requests of this window size is stored
const windowSize = time.Hour

func main() {
	fetch_content.Main() // Starting main function of Fetch Content

	// Used for shutting down workers
	shutdown = make(chan struct{}, MaxPaidWorkers+MaxPaidWorkers)

	paidJobQueue = make(chan Job, MaxPaidWorkers)
	nonPaidJobQueue = make(chan Job, MaxPaidWorkers)

	// Starting Workers
	for i := 1; i <= MaxPaidWorkers; i++ {
		wg.Add(1)
		go Worker(i, true)
	}

	for i := 1; i <= MaxNonPaidWorkers; i++ {
		wg.Add(1)
		go Worker(i, false)
	}

	//Starting server and routes
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/query", CrawlHandler)
	http.HandleFunc("/api/setCrawlers", setCrawlers)
	http.HandleFunc("/api/setSpeed", setSpeed)

	server_err := http.ListenAndServe(":3000", nil)
	if server_err != nil {
		fmt.Println("Unable to Start The server")
		return
	}
}

// To set the crawling speed
func setSpeed(w http.ResponseWriter, r *http.Request) {
	new_noOfReqStr := r.URL.Query().Get("speed")
	new_noOfReq, err := strconv.Atoi(new_noOfReqStr)
	if err != nil || new_noOfReq < 0 {
		fmt.Fprintf(w, "Invalid crawling speed")
		return
	}
	maxRequestsPerHour = new_noOfReq
	fmt.Fprintf(w, "Crawling speed set to %d", new_noOfReq)
}

// To handle url crawl requests
func CrawlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	isPayingCustomer := false
	isPayingCustomer = r.URL.Query().Get("isPaidCustomer") == "1"

	tries := 2
	if isPayingCustomer {
		tries++
	}
	job := Job{
		URL:            url,
		ResultChannel:  make(chan string),
		isPaidCustomer: isPayingCustomer,
		retries:        tries,
	}
	reqMutex.Lock()

	//Pushing into job queue
	if isPayingCustomer {
		paidJobQueue <- job
	} else {
		nonPaidJobQueue <- job
	}

	reqMutex.Unlock()
	result := <-job.ResultChannel // To obtain result from the result Channel
	job.retries--

	// Pushing the job into jobqueue again if fetching is failed, until a certain limit of retries
	if result == "" && job.retries > 0 {
		fmt.Println("Haaa")
		if isPayingCustomer {
			paidJobQueue <- job
		} else {
			nonPaidJobQueue <- job
		}
		result = <-job.ResultChannel
	}
	if result == "" {
		fmt.Fprintf(w, "Unable to fetch the given URL after %d tries", tries)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, result)
}

var workersChange sync.Mutex // Mutex to handle change in number of workers

func setCrawlers(w http.ResponseWriter, r *http.Request) {
	new_paidWorkers, err := strconv.Atoi(r.URL.Query().Get("paidWorkers"))
	new_unpaidWorkers, err1 := strconv.Atoi(r.URL.Query().Get("unpaidWorkers"))

	if err != nil || err1 != nil || new_paidWorkers <= 0 || new_unpaidWorkers <= 0 {
		// Error handling
		fmt.Println(err, err1)
		fmt.Fprintf(w, "Input error, Please set the value correctly")
		return
	}
	workersChange.Lock()
	defer workersChange.Unlock()
	initWorkers(new_paidWorkers, new_unpaidWorkers)
	fmt.Fprintf(w, "Set paid workers to %d and unpaidworkers to %d", new_paidWorkers, new_unpaidWorkers)
}

func Worker(id int, forPaid bool) {
	var requestTimes []time.Time // To maintain the count of requests in past hour
	var JobQueue chan Job

	// Setting Queue based on worker type
	if forPaid {
		JobQueue = paidJobQueue
	} else {
		JobQueue = nonPaidJobQueue
	}
	defer wg.Done()
	for {
		select {
		case <-shutdown:
			return //Signal for shutting down
		case job := <-JobQueue:
			currentTime := time.Now()

			//Removing requests which are older than 1 hour
			var i int
			for i = 0; i < len(requestTimes); i++ {
				if currentTime.Sub(requestTimes[i]) <= windowSize {
					break
				}
			}
			if i > 0 {
				requestTimes = requestTimes[i:]
			}
			// for len(requestTimes) > 0 && currentTime.Sub(requestTimes[0]) > windowSize {
			// 	requestTimes = requestTimes[1:]
			// }

			if len(requestTimes) >= maxRequestsPerHour {
				job.ResultChannel <- "Rate limit exceeded. Try again later."
				continue
			}
			requestTimes = append(requestTimes, currentTime) // Adding current request
			result, _ := fetch_content.FetchContent(job.URL, job.isPaidCustomer)
			job.ResultChannel <- result
		}
	}
}

// Re initialise workers to update the workers count
func initWorkers(new_paidWorkers, new_unpaidWorkers int) {

	// Signals for shutting down the workers
	for i := 1; i <= MaxNonPaidWorkers+MaxPaidWorkers; i++ {
		shutdown <- struct{}{}
	}
	wg.Wait()
	// After all the workers are down new workers are started
	for i := 1; i <= new_paidWorkers; i++ {
		wg.Add(1)
		go Worker(i, true)
	}

	for i := 1; i <= new_unpaidWorkers; i++ {
		wg.Add(1)
		go Worker(i, false)
	}

	MaxPaidWorkers = new_paidWorkers
	MaxNonPaidWorkers = new_unpaidWorkers
	newPaidJobQueue := make(chan Job, MaxPaidWorkers) // Update the queue size accordingly
	newNonPaidJobQueue := make(chan Job, MaxNonPaidWorkers)

	reqMutex.Lock() // Stop incoming requests to jobqueue to avoid race condition and make sure that tasks are not lost
	defer reqMutex.Unlock()

	// Shift all the jobs from old jobqueue to the new jobqueue
	for i := 0; i < len(paidJobQueue); {
		newPaidJobQueue <- <-paidJobQueue
	}
	for i := 0; i < len(nonPaidJobQueue); {
		newNonPaidJobQueue <- <-nonPaidJobQueue
	}
	paidJobQueue = newPaidJobQueue
	nonPaidJobQueue = newNonPaidJobQueue
	close(shutdown)
	shutdown = make(chan struct{}, MaxPaidWorkers+MaxPaidWorkers)
	// fmt.Printf("Number of paid workers set to %d and number of non-paid workers set to %d \n", MaxPaidWorkers, MaxNonPaidWorkers)
}
