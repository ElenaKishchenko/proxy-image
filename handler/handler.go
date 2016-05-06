package handler

import (
	"bytes"
	"crypto/md5"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"encoding/hex"
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

	rawMd5 := md5.Sum(this.imgBlob)
	nameMd5 := hex.EncodeToString(rawMd5[:])
	filename := "/tmp/" + nameMd5 + this.imgType
	fmt.Println(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("write file")
		err := ioutil.WriteFile(filename, this.imgBlob, 0644)
		if err != nil {
			return err
		}
		fmt.Println("resize")
		newImage, err := ResizeImg(this.imgBlob)
		if err != nil {
			return err
		}

		strUrl := SERVER2 + "?metod=file&filename=" + this.url //SERVERWWW + r.URL.String()

		_, err = http.Post(strUrl, this.cType, bytes.NewBuffer(newImage))
		if err != nil {
			return err
		}
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
