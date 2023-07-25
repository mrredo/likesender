package main

import (
	"bytes"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	Url    = "https://viesturi.edu.lv/wp-admin/admin-ajax.php"
	client = http.Client{}
	mapp   = app.New()
	win    = mapp.NewWindow("VVV viesturi like botter")
	stopCh = make(chan struct{})
	done   = make(chan bool)
	winX   = 500
	winY   = 700
)
var urlPost *widget.Entry
var times *widget.Entry
var Submit *widget.Button
var StopRequests *widget.Button
var PostIds *widget.Select

func main() {

	urlPost = widget.NewEntry()
	urlPost.PlaceHolder = "VVV viesturi post link..."
	times = widget.NewEntry()
	times.PlaceHolder = "How many likes to send...(type -1 for unlimited requests until stopped)"
	Submit = widget.NewButton("Send Requests", MakeRequest)
	StopRequests = widget.NewButton("Stop requests", StopRequest)
	PostIds = widget.NewSelect([]string{}, func(s string) {})
	PostIds.Disable()
	PostIds.PlaceHolder = "Select a post id"
	urlPost.OnSubmitted = UrlOnSubmit
	StopRequests.Disable()
	content := container.NewVBox(urlPost, times, PostIds, Submit, StopRequests)
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
	win.SetFixedSize(true)
	win.SetContent(content)
	win.ShowAndRun()
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
func LikePost(formData url.Values) {
	body := bytes.NewBufferString(formData.Encode())
	req, _ := http.NewRequest("POST", Url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	// Signal that the request is completed
	done <- true
}

func MakeRequest() {
	postid, err := ParseUrlToId(urlPost.Text)
	if err != nil {
		return
	}
	times, err := strconv.Atoi(times.Text)
	if err != nil {
		return
	}
	StopRequests.Importance = widget.DangerImportance
	Submit.Disable()
	StopRequests.Enable()

	// Use the 'done' channel to signal when the request handling is completed
	go func() {
		MultipleLikes(times, FormData(postid))
		done <- true
	}()
}

func StopRequest() {
	StopRequests.Importance = widget.MediumImportance
	Submit.Enable()
	StopRequests.Disable()

	// Signal to stop the requests
	stopCh <- struct{}{}

	// Wait for the request handling to complete
	<-done
}

func MultipleLikes(howmuch int, formData url.Values) {
	ticker := time.NewTicker(time.Millisecond * 400)
	defer ticker.Stop()
	if howmuch <= -1 {
		for {
			select {
			case <-stopCh:
				return // Stop if we receive a stop signal
			case <-ticker.C:
				go LikePost(formData)
			}

		}
	} else {
		for i := 0; i < howmuch; i++ {
			select {
			case <-stopCh:
				return // Stop if we receive a stop signal
			case <-ticker.C:
				go LikePost(formData)
			}

		}
	}

	stopCh <- struct{}{}
}

func PostIdsFromUrl(urlVVV string) {
	response, err := http.Get(urlVVV)
	if err != nil {
		urlPost.SetText("")
		PostIds.Disable()
		return
	}
	defer response.Body.Close()

	// Check if the request was successful (status code 200)
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch URL: %s returned status code %d", urlVVV, response.StatusCode)
	}

	// Read the response body into a string
	htmlContent, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		urlPost.SetText("")
		PostIds.Disable()
		return
	}
	options := []string{}
	htmlContent.Find(".mfn-love[data-id]").Each(func(_ int, s *goquery.Selection) {
		var ss, ok = s.Attr("data-id")
		if ok {
			options = append(options, ss)
		}
	})
	if len(options) == 0 {
		urlPost.SetText("")
		PostIds.Disable()
	} else {
		PostIds.Enable()
		PostIds.Options = options
		PostIds.Refresh()
	}
	return
}
