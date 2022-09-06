package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", inspectHandler)

	http.ListenAndServe(":3001", mux)
}

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s request started\n", r.Method, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		log.Printf("[%s] %s request finished. Code: %d\n", r.Method, r.URL.Path, http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		defer r.Body.Close()
		inputBody, err := ioutil.ReadAll(r.Body)
		var m InputModel
		if err = json.Unmarshal(inputBody, &m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			log.Printf("[%s] %s request finished. Code: %d. Reason: %s\n", r.Method, r.URL.Path, http.StatusBadRequest, err.Error())
			return
		}

		metas := make(map[string]*PageMeta)
		for _, url := range m.Urls {
			if meta, err := getPageMeta(url); err != nil {
				metas[*url] = nil
				log.Printf("Failed extract meta for %s. Reason: %s\n", *url, err.Error())
			} else {
				metas[*url] = meta
				log.Printf("Success extract meta for %s", *url)
			}
		}

		body, _ := json.Marshal(metas)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
		log.Printf("[%s] %s request finished. Code: %d\n", r.Method, r.URL.Path, http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func getPageMeta(url *string) (*PageMeta, error) {
	response, err := http.Get(*url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("Response code %d\n", response.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return extractMeta(doc), nil
}

func extractMeta(doc *goquery.Document) *PageMeta {
	title := doc.Find("title").Text()
	description, _ := doc.Find("meta[name=description]").Attr("content")

	ogTitle, _ := doc.Find("meta[property=\"og:title\"]").Attr("content")
	ogType, _ := doc.Find("meta[property=\"og:type\"]").Attr("content")
	ogImage, _ := doc.Find("meta[property=\"og:image\"]").Attr("content")
	ogUrl, _ := doc.Find("meta[property=\"og:url\"]").Attr("content")

	return &PageMeta{
		Title:       title,
		Description: description,
		OG: &OGMeta{
			Title: ogTitle,
			Type:  ogType,
			Image: ogImage,
			Url:   ogUrl,
		},
	}
}

type PageMeta struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	OG          *OGMeta `json:"og"`
}

type OGMeta struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Image string `json:"image"`
	Url   string `json:"url"`
}

type InputModel struct {
	Urls []*string `json:"urls"`
}
