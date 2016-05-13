package handler

import (
	"bytes"
	"gopkg.in/gographics/imagick.v2/imagick"
	"net/http"
	"crypto/md5"
	"encoding/hex"
	"os"
	"io/ioutil"
)

type Task struct {
	url     string
	imgBlob []byte
	imgType string
	method  string
	cType   string
	hAuth   string
	hCookie string
}

func (this *Task) Exec(serverUrl string, pathSave string, maxWidth uint) error {
	var newImage []byte
	rawMd5 := md5.Sum(this.imgBlob)
	nameMd5 := hex.EncodeToString(rawMd5[:])
	filename := pathSave + nameMd5 + this.imgType

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		newImage, err := ResizeImg(this.imgBlob, maxWidth)
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

	strUrl := serverUrl + this.url

	req, err := http.NewRequest(this.method, strUrl, bytes.NewBuffer(newImage))
	req.Header.Set("Authorization", this.hAuth)
	req.Header.Set("Content-Type", this.cType)
	req.Header.Set("Cookie", this.hCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func ResizeImg(imageData []byte, maxWidth uint) ([]byte, error) {
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

	if width > maxWidth {
		hWidth := maxWidth
		hHeight := uint(maxWidth * height / width)

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
