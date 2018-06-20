package server

import (
	d "github.com/dailyhunt/airdb/db"
	"github.com/dailyhunt/airdb/operation"
	"github.com/dailyhunt/airdb/utils"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var db d.DB

func StartHTTPServer(d d.DB, kvPort int) {
	db = d
	router := gin.New()
	//router.Use(gin.Logger())
	router.GET("/store/:key", getValue)
	router.POST("/store", setValue)
	router.POST("/raft/:id", addNode)
	adrr := ":" + strconv.Itoa(kvPort)
	logger.Debug("Starting rest api at  : ", adrr)
	router.Run(adrr)

}
func addNode(c *gin.Context) {
	id := c.Params.ByName("id")
	url, _ := ioutil.ReadAll(c.Request.Body)
	nodeId, _ := strconv.ParseInt(id, 0, 64)

	table, err := db.GetTable("t1")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		logger.Debug("Adding node ", nodeId, " with url ", string(url))
		err := table.AddRegionPeer(nodeId, url)
		// Todo(sohan) return response
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}

}

func getValue(c *gin.Context) {
	key := c.Params.ByName("key")
	c.JSON(200, gin.H{
		"key":   key,
		"value": key + "_value",
	})
}

func setValue(c *gin.Context) {
	var put operation.Put
	if err := c.ShouldBindJSON(&put); err == nil {

		if put.K == utils.EMPTY_STR || put.V == utils.EMPTY_STR {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content"})
			logger.WithFields(logger.Fields{
				"key":   put.K,
				"value": put.V,
			}).Error("Empty Key/Value POST Request ")
		} else {
			put.T = time.Now()
			logger.WithFields(logger.Fields{
				"key":   put.K,
				"value": put.V,
				"epoch": put.T,
			}).Info("KV POST Request ")

			table, err := db.GetTable("t1")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				err := table.Put(&put)
				// Todo(sohan) return response
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				}
			}

		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}
