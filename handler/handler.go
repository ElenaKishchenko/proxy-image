package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"net/http"
	"os"
)

const SERVER2 = "http://docker.local:8080"

type ProxyHandler struct {
	worksCount int32
	errors     []error
	Tasks      chan Task
}

type Task struct {
	url     string
	imgBlob []byte
	imgType string
	method  string
	cType   string
}

func (this *Task) Exec() error {
	var newImage []byte
	rawMd5 := md5.Sum(this.imgBlob)
	nameMd5 := hex.EncodeToString(rawMd5[:])
	filename := "/tmp/" + nameMd5 + this.imgType
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		newImage, err := ResizeImg(this.imgBlob)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, newImage, 0644)
		if err != nil {
			return err
		}

	}
	newImage, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	strUrl := SERVER2 + this.url
	fmt.Println(strUrl)
	_, err = http.Post(strUrl, this.cType, bytes.NewBuffer(newImage))
	if err != nil {
		return err
	}

	return nil
}

func ResizeImg(imageData []byte) ([]byte, error) {
	imagick.Initialize()
	defer imagick.Terminate()
	var err error
	mw := imagick.NewMagickWand()
	err = mw.ReadImageBlob(imageData)
	if err != nil {
		return nil, err
	}
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	if width > 1024 {
		hWidth := uint(1024)
		hHeight := uint(1024 * height / width)

		err = mw.ResizeImage(hWidth, hHeight, imagick.FILTER_LANCZOS, 1)
		if err != nil {
			return nil, err
		}
	}
	err = mw.SetImageCompressionQuality(90)
	if err != nil {
		return nil, err
	}

	res := mw.GetImageBlob()

	return res, nil

}
