package main

import (
	"bytes"
	"errors"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"net/http"
	"net/url"
	"strings"
)

var (
	Url    = "https://viesturi.edu.lv/wp-admin/admin-ajax.php"
	Action = "mfn_love"
	client = http.Client{}
	//method = POST
	//
	mapp = app.New()
	win  = mapp.NewWindow("Number guessing game using higher or lower")
)
var urlPost *widget.Entry
var times *widget.Entry
var Submit *widget.Button
var StopRequests *widget.Button

func main() {

	fmt.Println(ParseUrlToId("https://viesturi.edu.lv/17427-2/"))

	urlPost = widget.NewEntry()
	urlPost.PlaceHolder = "VVV viesturi post link..."
	times = widget.NewEntry()
	times.PlaceHolder = "How many likes to send...(type -1 for unlimited requests until stopped)"
	content := container.NewVBox(urlPost, times, Submit)

	win.SetContent()
	win.ShowAndRun()
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
	return strings.Split(url.Path[1:len(url.Path)-1], "-")[0], err
}
func LikePost(formData url.Values) {
	body := bytes.NewBufferString(formData.Encode())
	req, _ := http.NewRequest("POST", Url, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}
func MultipleLikes(howmuch int, formData url.Values, stopCh chan struct{}, done chan bool) {
	if howmuch <= -1 {
		for {
			select {
			case <-stopCh:
				return
			default:
				LikePost(formData)
			}
		}
	} else {
		for i := 0; i < howmuch; i++ {
			LikePost(formData)

		}
		done <- true
	}
}
