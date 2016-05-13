package main

import (
	"proxy-image/handler"
	"net/http"
	"gopkg.in/gographics/imagick.v2/imagick"
	"log"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Server string
	Path string
	MaxSizeImg uint
	MaxQueue int
}


func main() {
	imagick.Initialize()
	defer imagick.Terminate()

	config := LoadConfig("./config.json")

	proxy := new(handler.ProxyHandler)
	proxy.Server = config.Server
	proxy.PathTmp = config.Path
	proxy.ImgMaxWidht = config.MaxSizeImg

	go proxy.Setup(config.MaxQueue)

	http.HandleFunc("/",proxy.Handler)
	err := http.ListenAndServe(":8080", nil)
	log.Println(err)
}

func LoadConfig(path string) Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Config File Missing. ", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Println("Config Parse Error: ", err)
	}

	return config
}