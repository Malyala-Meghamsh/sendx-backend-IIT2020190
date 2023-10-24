package cache_manager

import (
	"io"
	"os"
	"strings"
	"sync"
)

var cachePath = "../Cache/"

func CreateDirectoryIfNotExist(directoryPath string) error {
	// Check if the directory already exists
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		// Directory does not exist, create it
		err := os.Mkdir(directoryPath, 0777)
		if err != nil {
			return err
		}
		return nil // Directory created successfully
	}
	return nil // Directory already exists
}

func getCacheFileName(url string) string {
	// Replace special characters in URL to create a valid file name
	// File names cannot be created having such characters
	url = strings.ReplaceAll(url, "\\", "_")
	url = strings.ReplaceAll(url, "|", "_")
	url = strings.ReplaceAll(url, ":", "_")
	url = strings.ReplaceAll(url, "\"", "_")
	url = strings.ReplaceAll(url, "?", "_")
	url = strings.ReplaceAll(url, "<", "_")
	url = strings.ReplaceAll(url, ">", "_")
	url = strings.ReplaceAll(url, "/", "_")
	return cachePath + url + ".txt"
}

func GetCachedContent(url string, cacheMutex *sync.RWMutex) (string, error) {
	// Obtain file name by given URL
	cacheFileName := getCacheFileName(url)
	file, err := os.Open(cacheFileName)
	if err != nil {
		cacheMutex.Unlock()
		// Because after that error goroutine execution
		// stops and if stops we may have deadlock because we never released lock
		return "", err
	}
	defer file.Close()

	content := make([]byte, 409600)
	// Read content from the cached file
	numberOfBytesRead, err := file.Read(content)
	if err != nil && err != io.EOF {
		cacheMutex.Unlock()
		return "", err
	}

	// Convert content to string and return
	return string(content[:numberOfBytesRead]), nil

}

func DeleteCachedContent(url string) error {
	cacheFileName := getCacheFileName(url)
	// Attempt to delete the file
	err := os.Remove(cacheFileName)
	if err != nil {
		return err
	}
	return nil // File deleted successfully
}

func StoreToCache(url, content string, cacheMutex *sync.RWMutex) error {
	cacheFileName := getCacheFileName(url)
	file, err := os.OpenFile(cacheFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		return err
	}

	// Create an io.Writer from the file descriptor
	writer := io.Writer(file)

	// Write content to the file using io.WriteString
	_, err = io.WriteString(writer, content)
	if err != nil {
		return err
	}

	// Content Stored in File
	return nil
}
