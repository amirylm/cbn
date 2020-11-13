package http

import (
	"github.com/amirylm/cbn/src/core"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func respond(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data, "time": time.Now().Unix()})
}

func RegisterBucketRoutes(router *gin.Engine, ctrl *core.Controller) error {
	// list all buckets
	router.GET("/buckets", func(c *gin.Context) {
		items := []interface{}{}
		ctrl.ListBuckets(func(bucket *core.Bucket) bool {
			msg := core.ToBucketMsg(bucket)
			items = append(items, msg)
			return true
		})
		respond(c, items)
	})

	// list bucket content (names)
	router.GET("/buckets/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		content, err := ctrl.GetBucketContent(hash)
		if err != nil {
			log.Panic("could not get bucket content")
		}
		respond(c, content)
	})

	// TODO: add api to update bucket source (to be called before upload of signed bucket)

	// upload signed bucket
	router.POST("/buckets/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		payload, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Panic("could not read payload")
		}
		bucket, err := core.ParseBucket(hash, payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not parse bucket"})
			return
		}
		err = ctrl.SaveSignedBucket(bucket)
		if err != nil {
			log.Panic("could not save bucket")
		}
		raw, err := core.SerializeBucket(bucket)
		if err != nil {
			log.Panic("could not serialize bucket")
		}
		respond(c, string(raw))
	})

	return nil
}