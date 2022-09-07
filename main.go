package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
)

var resultCache map[string]*PageMeta

func main() {
	resultCache = make(map[string]*PageMeta)

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

		outCh := make(chan *Result)
		for _, url := range m.Urls {

			go func(url *string, out chan<- *Result) {
				if meta, ok := resultCache[*url]; ok {
					outCh <- SuccessResult(url, meta)
					log.Printf("Meta extracted from cache for %s", *url)
				} else if doc, err := getPage(url); err != nil {
					out <- ErrorResult(url, err)
					log.Printf("Meta failed to extract %s. Reason: %s\n", *url, err.Error())
				} else {
					out <- SuccessResult(url, extractMeta(doc))
					log.Printf("Meta extracted for %s", *url)
				}
			}(url, outCh)
		}

		metas := make(map[string]*PageMeta)
		for range m.Urls {
			r := <-outCh
			if r.Error != nil {
			} else {
				resultCache[*r.Url] = r.Result
				metas[*r.Url] = r.Result
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

func getPage(url *string) (*goquery.Document, error) {
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

	return doc, nil
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

type Result struct {
	Error  error
	Url    *string
	Result *PageMeta
}

func ErrorResult(url *string, err error) *Result {
	return &Result{
		Url:   url,
		Error: err,
	}
}

func SuccessResult(url *string, meta *PageMeta) *Result {
	return &Result{
		Url:    url,
		Result: meta,
	}
}
