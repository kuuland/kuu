package kuu

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func isExecutable(data []byte) bool {
	elfMagicNumbers := []string{"7F", "45", "4C", "46"}

	for i, s := range data[:8] {
		if fmt.Sprintf("%X", s) != elfMagicNumbers[i%len(elfMagicNumbers)] {
			return false
		}
	}
	return true
}

func isText(data []byte) bool {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "text/html") {
		return true
	}
	return false
}

func detectXSS(input string) bool {
	var match = false
	xssRegex := regexp.MustCompile("<script.*?>|</script.*?>|<iframe.*?>|</iframe.*?>")
	match = xssRegex.MatchString(input)
	if strings.Contains(strings.ToLower(input), "javascript") {
		match = true
	}
	return match
}

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

func GetUploadDir() string {
	uploadDir := C().GetString("uploadDir")
	if uploadDir == "" {
		uploadDir = "assets/upload"
	}
	EnsureDir(uploadDir)
	return uploadDir
}

// SaveUploadedFile
func SaveUploadedFile(fh *multipart.FileHeader, save2db bool, extraData ...*File) (f *File, err error) {
	uploadDir := GetUploadDir()

	src, err := fh.Open()
	if err != nil {
		return f, err
	}
	defer func() {
		if err := src.Close(); err != nil {
			ERROR(err)
		}
	}()
	body, err := io.ReadAll(src)
	if err != nil {
		return f, err
	}

	if C().DefaultGetBool("upload.EnableExe", false) && isExecutable(body) {
		return f, fmt.Errorf("file is executable")
	}
	if C().DefaultGetBool("upload.EnableXSS", false) && isText(body) {
		if detectXSS(string(body)) {
			return f, fmt.Errorf("xss detected")
		}
	}

	md5Sum := fmt.Sprintf("%x", md5.Sum(body))

	// 保存文件
	md5Name := fmt.Sprintf("%s%s", md5Sum, path.Ext(fh.Filename))
	dst := path.Join(uploadDir, md5Name)
	if err := saveFile(fh, dst); err != nil {
		return f, err
	}

	f = &File{
		UID:  strings.ReplaceAll(uuid.NewV4().String(), "-", ""),
		Type: fh.Header["Content-Type"][0],
		Size: fh.Size,
		Name: fh.Filename,
		URL:  fmt.Sprintf("/%s/%s", strings.Trim(uploadDir, "/"), md5Name),
		Path: dst,
	}

	if len(extraData) > 0 && extraData[0] != nil {
		extra := extraData[0]
		f.Class = extra.Class
		f.OwnerType = extra.OwnerType
		f.OwnerID = extra.OwnerID
		f.ExtendField = extra.ExtendField
		f.Model = extra.Model
	}
	if save2db {
		if err := DB().Create(&f).Error; err != nil {
			return nil, err
		}
	}
	return
}
