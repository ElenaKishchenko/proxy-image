package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
	"bytes"
	"errors"
	"path/filepath"
)

type ProxyHandler struct {
	worksCount int32
	Server string
	PathTmp string
	ImgMaxWidht uint
	errors     []error
	Tasks      chan Task
}

func (this *ProxyHandler) Setup(maxWorker int) error {
	results := make(chan error, maxWorker)
	this.Tasks = make(chan Task, maxWorker)

	for i := 0; i < maxWorker; i++ {
		go func() {
			for task := range this.Tasks {
				results <- task.Exec(this.Server,this.PathTmp,this.ImgMaxWidht)
			}
		}()
	}

	go func() {
		for result := range results {
			if result != nil {
				this.errors = append(this.errors, result)
			}
			atomic.AddInt32(&this.worksCount, -1) //Atomic decrement
		}
	}()

	return nil

}

func (this *ProxyHandler) Handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Query().Get("mode") == "import" {
		for atomic.LoadInt32(&this.worksCount) > 0 {
			time.Sleep(time.Second)
		}

		if len(this.errors) > 0 {
			var resErrors string
			for _, errStr := range this.errors {
				resErrors += errStr.Error() + "\n"
			}
			w.Write([]byte(resErrors))
		} else {
			sendReturn(w, r, this.Server)
		}

	} else if r.URL.Query().Get("mode") == "file" && filepath.Ext(r.URL.Query().Get("filename")) != ".xml" {

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, err)
				return
			}
			var imgType string
			switch http.DetectContentType(body) {
			case "image/jpeg":
				imgType = ".jpg"
			case "image/png":
				imgType = ".png"
			case "image/gif":
				imgType = ".gif"
			default:
				err := errors.New("Не корректный тип файла " + r.URL.Query().Get("filename"))
				writeError(w, err)
				return
			}
			var task Task
			task.url = r.URL.String()
			task.imgBlob = body
			task.imgType = imgType
			task.method = r.Method
			task.cType = r.Header.Get("Content-Type")
			task.hAuth = r.Header.Get("Authorization")
			task.hCookie = r.Header.Get("Cookie")

			atomic.AddInt32(&this.worksCount, 1) // Atomic increment
			this.Tasks <- task

			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("success\n"))

	} else {

		sendReturn(w, r, this.Server)
	}
}

func writeError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(err.Error()))
}

func sendReturn(w http.ResponseWriter, r *http.Request, serverUrl string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	req, err := http.NewRequest(r.Method, serverUrl + r.URL.String(), bytes.NewBuffer(body))
	//req.Header = r.Header
	req.Header.Set("Authorization", r.Header.Get("Authorization"))
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	req.Header.Set("Cookie", r.Header.Get("Cookie"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, err)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Write(body)
}
