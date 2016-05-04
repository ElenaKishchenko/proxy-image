package main

import (
	"fmt"
	"proxy-image/handler"
	"net/http"
	"gopkg.in/gographics/imagick.v2/imagick"
)



func main() {
	imagick.Initialize()

	http.HandleFunc("/", handler.Handler)
	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}
