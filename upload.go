package kuu

import (
	"crypto/md5"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

func saveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// SaveUploadedFile
func SaveUploadedFile(fh *multipart.FileHeader, save2db bool, extraData ...*File) (f *File, err error) {
	uploadDir := C().GetString("uploadDir")
	uploadPrefix := C().GetString("uploadPrefix")
	if uploadDir == "" {
		uploadDir = "assets/upload"
	}
	if uploadPrefix == "" {
		uploadPrefix = uploadDir
	}
	EnsureDir(uploadDir)

	src, err := fh.Open()
	if err != nil {
		return f, err
	}
	defer func() {
		if err := src.Close(); err != nil {
			ERROR(err)
		}
	}()
	body, err := ioutil.ReadAll(src)
	md5Sum := fmt.Sprintf("%x", md5.Sum(body))

	//保存文件
	md5Name := fmt.Sprintf("%s%s", md5Sum, path.Ext(fh.Filename))
	dst := path.Join(uploadDir, md5Name)
	if err := saveFile(fh, dst); err != nil {
		return f, err
	}

	f = &File{
		UID:    uuid.NewV4().String(),
		Type:   fh.Header["Content-Type"][0],
		Size:   fh.Size,
		Name:   fh.Filename,
		Status: "done",
		URL:    fmt.Sprintf("/%s/%s", strings.Trim(uploadPrefix, "/"), md5Name),
		Path:   dst,
	}

	if len(extraData) > 0 && extraData[0] != nil {
		extra := extraData[0]
		f.RefID = extra.RefID
		f.Class = extra.Class
	}
	if save2db {
		if err := DB().Create(&f).Error; err != nil {
			return nil, err
		}
	}
	return
}
