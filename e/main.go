package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Replace 'url' with the actual URL of the webpage you want to scrape
	url := "https://viesturi.edu.lv/"

	// Fetch the HTML content from the URL
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Check if the request was successful (status code 200)
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch URL: %s returned status code %d", url, response.StatusCode)
	}

	// Read the response body into a string
	htmlContent, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(htmlContent.Find(".mfn-love").Text())

	// Find all elements with class "mfn_love" and "data-id" attribute
	htmlContent.Find(".mfn-love[data-id]").Each(func(_ int, s *goquery.Selection) {
		// Now you can work with the found element
		fmt.Println(s.Attr("data-id")) // Prints the text content of the element
	})
}
