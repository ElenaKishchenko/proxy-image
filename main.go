package main

import (
	"fmt"
	"proxy-image/handler"
	"net/http"
	"io"
	"bytes"
	"strconv"
	"mime/multipart"
)



func main() {

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}
