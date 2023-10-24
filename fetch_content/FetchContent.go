package fetch_content

import (
	"fmt"
	"net/http"
	cachePackage "sendx/cache_manager"
	"sendx/error_handler"
	"strings"

	// "sendx/main"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Maintain map of URLs crawled in in-memory and store the files actually in disk

type URLCache struct {
	mu  sync.RWMutex         // Mutex to prevent race conditions
	url map[string]time.Time // Map of URLs crawled and time
}

// Creating URL Cache
func NewURLCache() *URLCache {
	return &URLCache{
		url: make(map[string]time.Time),
	}
}

// Adding an URL
func (c *URLCache) Add(new_url string, t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.url[new_url] = t
}

var cache *URLCache

// var paidWorkersMutex sync.RWMutex
// var unpaidWorkersMutex sync.RWMutex

// Directory to store all the fetched pages if found will be returned from disk
const cacheDirectoryPath string = "../Cache"

func Main() {
	cache = NewURLCache()
	go startCacheCleanup(cache)
	err_direcCreation := cachePackage.CreateDirectoryIfNotExist(cacheDirectoryPath)
	error_handler.LogError(err_direcCreation, "Cannot Create Directory", nil)
	if err_direcCreation != nil {
		return
	}
}

func FetchContent(inputURL string, isPaidCustomer bool) (string, error) {
	cache.mu.Lock()
	if cachedResponse_time, found := cache.url[inputURL]; found && time.Now().Sub(cachedResponse_time).Minutes() <= 60 {
		// Fetch the URL Page from disk as it was fetched in last 60 minutes

		defer cache.mu.Unlock()
		cachedContent, err_getCachedContent := cachePackage.GetCachedContent(inputURL, &(cache.mu))
		error_handler.LogError(err_getCachedContent, "Unable to get cached content of url : "+inputURL, nil)
		if err_getCachedContent != nil {
			// If unable to fetch from file
			return "", err_getCachedContent
		}
		// fmt.Println("Value Found No need To Compute")
		return cachedContent, nil
	} else {
		// Crawling the URL using fetch function
		cache.mu.Unlock()

		fetchedData, err_FetchingData := fetchURL(inputURL, isPaidCustomer)

		error_handler.LogError(err_FetchingData, "Unable to fetch the given URL : "+inputURL, nil)
		if err_FetchingData != nil {
			return "", err_FetchingData //Error handling
		}

		cache.mu.Lock()
		defer cache.mu.Unlock()

		// Saving the file crawled in disk and inmemory URLcache
		cache.url[inputURL] = time.Now()
		err_storingData := cachePackage.StoreToCache(inputURL, fetchedData, &cache.mu)
		error_handler.LogError(err_storingData, "Cannot store it in cache", nil)
		if err_storingData != nil {
			return fetchedData, err_storingData // Fetched Data will be returned but we are unable to store it in file
		}

		// fmt.Println("Calculated new value")
		return fetchedData, nil
	}
}

// Cleanup every 30 minutes to maintain cache space
func startCacheCleanup(cache *URLCache) {
	cleanupInterval := 30 * time.Minute
	for {
		time.Sleep(cleanupInterval) // Clean up at 30 minutes interval
		cache.mu.Lock()
		defer cache.mu.Unlock()
		for url, timestamp := range cache.url {
			if time.Since(timestamp) > 60*time.Minute {
				// Erased in both inmemory URLCache and disk
				delete(cache.url, url)
				err_deleteFile := cachePackage.DeleteCachedContent(url)
				error_handler.LogError(err_deleteFile, "Cannot Delete file "+url, nil)
				//Will be continuing tasks, but will be alerting the admin through logging in terminal
			}
		}
	}
}

func fetchURL(inputURL string, isPaidCustomer bool) (string, error) {
	var retries int

	if isPaidCustomer {
		retries = 3
	} else {
		retries = 2
	}
	// Will be Retrying until our crawling is successful
	for i := 0; i < retries; i++ {
		// res_http, err_httpGet := http.Get(inputURL)
		// if err_httpGet != nil {
		// 	continue
		// }
		// defer res_http.Body.Close()
		// body_http, err_httpRead := io.ReadAll(res_http.Body)
		// if err_httpRead != nil {
		// 	continue
		// }
		// return string(body_http), nil
		s, err := crawlsite(inputURL)
		if err != nil {
			continue
		}
		return s, nil
	}
	return "", fmt.Errorf("Unable to fetch URL after " + strconv.Itoa(retries) + " retries")
}

// Crawling the URL
func crawlsite(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	// Find and accumulate all links on the webpage
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			result.WriteString(fmt.Sprintf("Link: %s\n", link))
		}
	})

	// Find and accumulate all headings on the webpage
	doc.Find(":header").Each(func(i int, s *goquery.Selection) {
		result.WriteString(fmt.Sprintf("Heading: %s\n", s.Text()))
	})

	return result.String(), nil
}
