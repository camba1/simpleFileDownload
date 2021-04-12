package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

// result: contains a string representation of a file contents and, optionally, an error
type result struct {
	fileContent string
	err         error
}

// processBody returns a representation of the contents of the file
// depending on the content type received
func processBody(contentType string, bBody []byte) (string, error) {
	returnval := ""
	switch contentType {
	case "text/plain":
		returnval = string(bBody)
	case "application/pdf":
		returnval = "got PDF file"
	case "application/json":
		var interBody interface{}
		err := json.Unmarshal(bBody, &interBody)
		data := interBody.(map[string]interface{})
		if err != nil {
			return "", fmt.Errorf("unable to unmarshall . Error: %v", err)
		}
		returnval = fmt.Sprint(data)
	default:
		return "", fmt.Errorf("unexpected response type: %s", contentType)
	}
	return returnval, nil
}

// downloadFile: Download a file and return a string representation of its contents
func downloadFile(wg *sync.WaitGroup, outputChannel chan result, url string) {

	defer wg.Done()

	// Download file

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		outputChannel <- result{
			fileContent: "",
			err:         fmt.Errorf("unable create new request. Error: %v", err),
		}
		return
	}
	req.Header.Add("User-Agent", "BBllc")

	resp, err := client.Do(req)
	if err != nil {
		outputChannel <- result{
			fileContent: "",
			err:         fmt.Errorf("unable to send  request. Error: %v", err),
		}
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	bBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		outputChannel <- result{
			fileContent: "",
			err:         fmt.Errorf("no  data received. Error: %v", err),
		}
		return
	}

	//Process body of response (file)
	returnVal, err := processBody(resp.Header.Get("Content-Type"), bBody)

	// Return results
	outputChannel <- result{
		fileContent: returnVal,
		err:         err,
	}
}

// readFilefromUrl:  download and process an individual file
func readFilefromUrl(wg *sync.WaitGroup, url string) {
	outputChannel := make(chan result)
	wg.Add(1)
	go downloadFile(wg, outputChannel, url)
	readResult := <-outputChannel
	if readResult.err != nil {
		log.Printf("error reading file: %v\n", readResult.err)
		return
	}
	log.Printf("result: \n%s\n", readResult.fileContent)
}

// dupUrl checks to see if we already processed this url
func dupUrl(urlList []string, url string) bool {
	for _, urlVisited := range urlList {
		if url == urlVisited {
			return true
		}
	}
	return false
}

// getUniqueUrlList: Returns a unique list of urls
func getUniqueUrlList(urls []string) []string {
	uniqueUrlList := make([]string, 0)
	for _, url := range urls {
		if !dupUrl(uniqueUrlList, url) {
			uniqueUrlList = append(uniqueUrlList, url)
		}
	}
	return uniqueUrlList
}

// readFilesFromUrls download anf process the files from the urls provided
func readFilesFromUrls(urls []string) {
	wg := sync.WaitGroup{}
	uniqueUrlList := getUniqueUrlList(urls)
	for _, url := range uniqueUrlList {
		readFilefromUrl(&wg, url)
	}
}

// openFile: Open text file and return individual lines as a string slice
func openFile(filename string) ([]string, error) {
	fileByte, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to Read file")
	}
	fileString := string(fileByte)
	fileLines := strings.Split(fileString, "\n")
	return fileLines, nil
}

func main() {
	//Get list of urls from file
	urls, err := openFile("test.txt")
	if err != nil {
		log.Fatal("unable to get urls list")
	}

	// Download and read files from urls
	readFilesFromUrls(urls)
}
