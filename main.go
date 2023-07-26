package main

import (
	"bytes"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	Url     = "https://viesturi.edu.lv/wp-admin/admin-ajax.php"
	client  = http.Client{}
	mapp    = app.New()
	win     = mapp.NewWindow("VVV viesturi like botter")
	stopCh  = make(chan struct{})
	done    = make(chan bool)
	likeSen = make(chan struct{})
	winX    = 500
	winY    = 700
)
var urlPost *widget.Entry
var urlSubmitUrl *widget.Button
var times *widget.Entry
var Submit *widget.Button
var StopRequests *widget.Button
var PostIds *widget.Select
var InfoAboutStuff *widget.Label

func main() {

	urlPost = widget.NewEntry()
	urlPost.PlaceHolder = "VVV viesturi post link..."
	urlPost.SetText("https://viesturi.edu.lv/")
	times = widget.NewEntry()
	times.PlaceHolder = "How many likes to send...(type -1 for unlimited requests until stopped)"
	Submit = widget.NewButton("Send Requests", MakeRequest)
	urlSubmitUrl = widget.NewButton("Submit url", func() {
		UrlOnSubmit(urlPost.Text)
	})
	InfoAboutStuff = widget.NewLabel("")

	StopRequests = widget.NewButton("Stop requests", StopRequest)
	PostIds = widget.NewSelect([]string{}, func(s string) {})
	PostIds.Disable()
	PostIds.PlaceHolder = "Select a post id"
	urlPost.OnSubmitted = UrlOnSubmit

	StopRequests.Disable()
	content := container.NewVBox(urlPost, urlSubmitUrl, times, PostIds, Submit, StopRequests, InfoAboutStuff)
	go func() {
		for {
			select {
			case <-stopCh:
				StopRequests.Disable()
				Submit.Enable()

			default:

			}
		}
	}()
	win.Resize(fyne.Size{
		Width:  float32(winX),
		Height: float32(winY),
	})
	win.CenterOnScreen()
	//win.SetFixedSize(true)
	win.SetContent(content)
	win.ShowAndRun()
}
func InfoLabel(likes int) {
	InfoAboutStuff.SetText(fmt.Sprintf(`
	url: %s
	current likes: %d
`, urlPost.Text, likes))
	//add a time how long
}
func UrlOnSubmit(s string) {
	PostIdsFromUrl(s)
}
func FormData(postid string) url.Values {
	formData := url.Values{}
	formData.Set("action", "mfn_love")
	formData.Set("post_id", postid)
	return formData
}
func ParseUrlToId(url1 string) (id string, err error) {
	url, err := url.Parse(url1)
	if len(url.Path) < 4 {
		return "", errors.New("invalid path for url")
	}
	url1path := strings.Split(url.Path[1:], "/")[0]
	return strings.Split(url1path, "-")[0], err
}
func LikePost(formData url.Values, likes *int) {
	body := bytes.NewBufferString(formData.Encode())
	req, _ := http.NewRequest("POST", Url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return
	}

	// Convert the response body to an integer
	intValue, err := strconv.Atoi(string(responseBody))
	if err != nil {
		return
	}
	if err == nil {
		*likes = intValue
	}

	done <- true
}

func MakeRequest() {
	times, err := strconv.Atoi(times.Text)
	if err != nil {
		return
	}
	StopRequests.Importance = widget.DangerImportance
	Submit.Disable()
	StopRequests.Enable()
	selected := strings.Replace(strings.Split(PostIds.Selected, "-")[1], " ", "", -1)
	// Use the 'done' channel to signal when the request handling is completed
	go func() {
		MultipleLikes(times, FormData(selected))
		done <- true
	}()
}

func StopRequest() {
	StopRequests.Importance = widget.MediumImportance
	Submit.Enable()
	StopRequests.Disable()
	PostIds.Enable()

	// Signal to stop the requests
	stopCh <- struct{}{}

	// Wait for the request handling to complete
	<-done
}

func MultipleLikes(howmuch int, formData url.Values) {
	ticker := time.NewTicker(time.Millisecond * 400)
	defer ticker.Stop()
	var likes = 0
	if howmuch <= -1 {
		for {
			select {
			case <-stopCh:
				return // Stop if we receive a stop signal
			case <-ticker.C:
				go LikePost(formData, &likes)
				InfoLabel(likes)
			}

		}
	} else {
		for i := 0; i < howmuch; i++ {
			select {
			case <-stopCh:
				return // Stop if we receive a stop signal
			case <-ticker.C:
				go LikePost(formData, &likes)
				InfoLabel(likes)
			}

		}
	}

	stopCh <- struct{}{}
}

func PostIdsFromUrl(urlVVV string) {
	response, err := http.Get(urlVVV)
	if err != nil {
		urlPost.SetText("https://viesturi.edu.lv/")
		PostIds.Disable()
		return
	}
	defer response.Body.Close()

	// Check if the request was successful (status code 200)
	if response.StatusCode != http.StatusOK {
		//	log.Fatalf("Failed to fetch URL: %s returned status code %d", urlVVV, response.StatusCode)
		urlPost.SetText("https://viesturi.edu.lv/")
		PostIds.Disable()
		return
	}

	// Read the response body into a string
	htmlContent, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		urlPost.SetText("https://viesturi.edu.lv/")
		PostIds.Disable()
		return
	}
	options := []string{}
	optionsTitles := []string{}
	sel := htmlContent.Find(".entry-title")
	if sel.Length() != 0 {
		sel.Each(func(_ int, s *goquery.Selection) {
			optionsTitles = append(optionsTitles, s.Find("a").Text())
		})
	} else {
		htmlContent.Find("h1").Each(func(_ int, s *goquery.Selection) {
			optionsTitles = append(optionsTitles, s.Text())
		})
	}

	htmlContent.Find(".mfn-love[data-id]").Each(func(_ int, s *goquery.Selection) {
		var ss, ok = s.Attr("data-id")

		if ok {
			options = append(options, ss)
		}
	})
	if len(options) == 0 {
		urlPost.SetText("https://viesturi.edu.lv/")
		PostIds.Disable()
	} else {
		combinedOptions := combineStringLists(options, optionsTitles)
		PostIds.Enable()
		PostIds.Options = combinedOptions
		PostIds.Refresh()
		PostIds.SetSelectedIndex(0)
	}
	return
}
func combineStringLists(list1, list2 []string) []string {
	combined := make([]string, 0, len(list1))
	for i := 0; i < len(list1); i++ {
		combined = append(combined, list2[i]+" - "+list1[i])
	}

	return combined
}
