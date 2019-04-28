package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/sys/models"
	"github.com/satori/go.uuid"
)

// Upload 文件上传
func Upload() kuu.RouteInfo {
	uploadDir := kuu.Join(kuu.ROOT, "/assets/")
	kuu.EnsureDir(uploadDir)

	return kuu.RouteInfo{
		Method: "POST",
		Path:   "/upload",
		Handler: func(c *gin.Context) {
			file, _ := c.FormFile("file")
			dst := kuu.Join(uploadDir, file.Filename)
			c.SaveUploadedFile(file, dst)
			kuu.Info(fmt.Sprintf("'%s' uploaded!", dst))

			File := kuu.Model("File")
			f := &models.File{
				UID:    uuid.NewV4().String(),
				Type:   file.Header["Content-Type"][0],
				Size:   file.Size,
				Name:   file.Filename,
				Status: "done",
				URL:    kuu.Join("/assets/", file.Filename),
				Path:   dst,
			}

			ret, err := File.Create(f)
			if err == nil {
				c.JSON(http.StatusOK, kuu.StdOK(ret[0]))
			} else {
				c.JSON(http.StatusOK, kuu.StdError(kuu.L(c, "upload_failed")))
			}
		},
	}
}
