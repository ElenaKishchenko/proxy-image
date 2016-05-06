package main

import (
	"proxy-image/handler"
	"net/http"
	"gopkg.in/gographics/imagick.v2/imagick"
	"log"
)


func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	max_Queue := 5

	proxy := new(handler.ProxyHandler)
	go proxy.Setup(max_Queue)

	http.HandleFunc("/",proxy.Handler)
	err := http.ListenAndServe(":8080", nil)
	log.Println(err)
}
