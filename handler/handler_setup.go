package handler

import (
	"net/http"
	"sync/atomic"
	"time"
	"io/ioutil"
	"errors"
	"fmt"
	"strings"
)

func (this *ProxyHandler) Setup(maxWorker int) error {
	results := make(chan error, maxWorker)
	this.Tasks = make(chan Task, maxWorker)

	for i := 0; i < maxWorker; i++ {
		go func() {
			for task := range this.Tasks {
				results <- task.Exec()
			}
		}()
	}

	go func () {
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
			for _,errStr :=  range this.errors {
				resErrors += errStr.Error() + "\n"
			}
			w.Write([]byte(resErrors))
		} else {
			strUrl := SERVER2 + r.URL.String()

			_, err := http.Post(strUrl, r.Header.Get("Content-Type"), r.Body)
			if err != nil {
				writeError(w, err)
				return
			}

			w.Write([]byte("success"))
		}
	} else if r.URL.Query().Get("mode") == "file" {

		fmt.Println("Content-Type = " + r.Header.Get("Content-Type"))
		//fmt.Println(r.Header.Get("Content-Lenght"))

		var imgType string
		switch strings.ToLower(r.Header.Get("Content-Type")) {
		case "image/jpeg":
			imgType = ".jpg"
		case "image/png":
			imgType = ".png"
		case "image/gif":
			imgType = ".gif"
		default:
			err := errors.New("No image")
			writeError(w, err)
			return
		}

		if r.ContentLength == 0 {
			err := errors.New("Body empty")
			writeError(w, err)
			return
		} else {
			body, err := ioutil.ReadAll(r.Body) //r.Body
			if err != nil {
				writeError(w, err)
				return
			}

			var task Task
			task.url = r.URL.String()
			task.imgBlob = body
			task.imgType = imgType
			task.method = r.Method
			task.cType = r.Header.Get("Content-Type")

			atomic.AddInt32(&this.worksCount, 1) // Atomic increment
			this.Tasks <- task
		}
	}


}

func writeError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.Header().Set("Content-Type","text/plain")
	w.Write([]byte(err.Error()))
}