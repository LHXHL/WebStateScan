package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
)

func main() {
	url := "https://bilibili.com/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//fmt.Println(resp.Body)
	//doc, err := goquery.NewDocumentFromReader(resp.Body)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(doc.Text())
	//fmt.Println(resp.Body)

	//var contentLength int64
	if resp.Header.Get("Content-Length") != "" {

	} else {
		fmt.Println(resp.Body)

		buf := new(bytes.Buffer)
		teeReader := io.TeeReader(resp.Body, buf)
		doc, err := goquery.NewDocumentFromReader(teeReader)
		if err != nil {
			return
		}
		fmt.Println(doc.Text())
		content := buf.Bytes()
		fmt.Println(len(content))
	}

	//head := doc.Find("head")
	//title := head.Find("title").Text()
	//fmt.Println(title)
}
