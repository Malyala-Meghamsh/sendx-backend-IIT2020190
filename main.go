package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	cachePackage "sendx/cache_manager"
	"sendx/error_handler"

	"github.com/gorilla/mux"
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

func startCacheCleanup(cache *URLCache) {
	cleanupInterval := 30 * time.Minute // Cleanup every 1 hour (adjust as needed)
	for {
		time.Sleep(cleanupInterval)
		cache.mu.Lock()
		for url, timestamp := range cache.url {
			if time.Since(timestamp) > 60*time.Minute {
				delete(cache.url, url)
				err_deleteFile := cachePackage.DeleteCachedContent(url)
				error_handler.LogError(err_deleteFile, "Cannot Delete file "+url, nil)
			}
		}
		cache.mu.Unlock()
	}
}

var cache *URLCache

var paidWorkerCount = 5
var workerPaidSemaphore = make(chan struct{}, paidWorkerCount)

var unpaidWorkersCount = 2
var workerUnpaidSemaphore = make(chan struct{}, paidWorkerCount)

const cacheDirectoryPath string = "../Cache"

func main() {
	cache = NewURLCache()
	go startCacheCleanup(cache)
	router := mux.NewRouter()
	err_direcCreation := cachePackage.CreateDirectoryIfNotExist(cacheDirectoryPath)
	error_handler.LogError(err_direcCreation, "Cannot Create Directory", nil)
	router.HandleFunc("/query/", serve).Methods("GET")
	// routes.Main()
	http.Handle("/", router)
	err_startServer := http.ListenAndServe(":3000", nil)
	error_handler.LogError(err_startServer, "Cannot Start Server", nil)
}

func serve(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	isPaidCustomer := false
	isPaidCustomer = r.URL.Query().Get("isPaidCustomer") == "1"
	fmt.Println("Hiii", isPaidCustomer)
	if url == "" {
		error_handler.LogError(fmt.Errorf("Missing URL Parameter"), "URL not found in query", w)
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	var wg sync.WaitGroup
	// fmt.Println(query["url"][0], "Wait", time.Now())
	if isPaidCustomer {
		workerPaidSemaphore <- struct{}{}
	} else {
		workerUnpaidSemaphore <- struct{}{}
	}
	// fmt.Println(query["url"][0], "Done Waiting", time.Now())
	wg.Add(1)
	go func(wr http.ResponseWriter) {
		cache.mu.Lock()
		defer func() {
			if isPaidCustomer {
				<-workerPaidSemaphore
			} else {
				<-workerUnpaidSemaphore
			}
			wg.Done()
		}()
		inputURL := query["url"][0]
		if cachedResponse_time, found := cache.url[inputURL]; found && time.Now().Sub(cachedResponse_time).Minutes() <= 60 {
			defer cache.mu.Unlock()
			cachedContent, err_getCachedContent := cachePackage.GetCachedContent(inputURL, &(cache.mu))
			error_handler.LogError(err_getCachedContent, "Unable to get cached content of url : "+inputURL, wr)
			fmt.Fprintf(wr, cachedContent)
			fmt.Println("Value Found No need To Compute")
		} else {
			cache.mu.Unlock()
			fetchedData, err_FetchingData := fetchURL(inputURL, isPaidCustomer)
			error_handler.LogError(err_FetchingData, "Unable to fetch the given URL : "+inputURL, wr)
			cache.mu.Lock()
			cache.url[inputURL] = time.Now()
			err_storingData := cachePackage.StoreToCache(inputURL, fetchedData, &cache.mu)
			error_handler.LogError(err_storingData, "Cannot store it in cache", nil)
			cache.mu.Unlock()
			fmt.Fprintf(wr, fetchedData)
			fmt.Println("Calculated new value")
		}
	}(w)
	wg.Wait()
}

func fetchURL(inputURL string, isPaidCustomer bool) (string, error) {
	var retries int
	if isPaidCustomer {
		retries = 5
	} else {
		retries = 3
	}
	for i := 0; i < retries; i++ {
		res_http, err_httpGet := http.Get(inputURL)
		if err_httpGet != nil {
			continue
		}
		defer res_http.Body.Close()
		body_http, err_httpRead := io.ReadAll(res_http.Body)
		if err_httpRead != nil {
			continue
		}
		return string(body_http), nil
	}
	return "", fmt.Errorf("Unable to fetch URL")
}
