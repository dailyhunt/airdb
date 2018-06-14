package server

import (
	"github.com/dailyhunt/airdb/utils"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func StartHTTPServer() {
	router := gin.New()
	//router.Use(gin.Logger())
	router.GET("/store/:key", getValue)
	router.POST("/store", setValue)
	router.Run(":8080")

}

func getValue(c *gin.Context) {
	key := c.Params.ByName("key")
	c.JSON(200, gin.H{
		"key":   key,
		"value": key + "_value",
	})
}

func setValue(c *gin.Context) {
	var kvCommand KvCommand
	if err := c.ShouldBindJSON(&kvCommand); err == nil {

		if kvCommand.Key == utils.EMPTY_STR || kvCommand.Value == utils.EMPTY_STR {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content"})
			logger.WithFields(logger.Fields{
				"key":   kvCommand.Key,
				"value": kvCommand.Value,
			}).Error("Empty Key/Value POST Request ")
		} else {
			kvCommand.Epoch = time.Now()
			kvCommand.Op = OP_SET
			logger.WithFields(logger.Fields{
				"key":   kvCommand.Key,
				"value": kvCommand.Value,
				"op":    "SET",
				"epoch": kvCommand.Epoch,
			}).Info("KV POST Request ")

		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}
