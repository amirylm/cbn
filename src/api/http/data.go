package http

import (
	"github.com/amirylm/cbn/src/core"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func RegisterDownloadRoutes(router *gin.Engine, ctrl *core.Controller) error {
	// download data from some bucket
	router.GET("/buckets/:hash/:name", func(c *gin.Context) {
		hash := c.Param("hash")
		name := c.Param("name")

		reader, ref,  err := ctrl.Download(hash, name)
		if err != nil {
			log.Panic("could not download data:", err)
		}

		c.DataFromReader(http.StatusOK, int64(ref.Header.Size), ref.Header.Type, reader, map[string]string{})
	})
	return nil
}

func RegisterUploadRoutes(router *gin.Engine, ctrl *core.Controller) error {

	// upload a file
	router.POST("/file", func(c *gin.Context) {
		file, fh, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not read file"})
			return
		}
		dr, err := ctrl.UploadData(*core.NewFileHeader(fh.Filename, fh.Header.Get("Content-Type")), file)
		if err != nil {
			log.Panic("could not upload file")
		}
		respond(c, dr)
	})

	// upload any payload
	router.POST("/data/:name", func(c *gin.Context) {
		name := c.Param("name")
		dr, err := ctrl.UploadData(*core.NewFileHeader(name, c.ContentType()), c.Request.Body)
		if err != nil {
			log.Panic("could not upload data")
		}
		respond(c, dr)
	})

	return nil
}
