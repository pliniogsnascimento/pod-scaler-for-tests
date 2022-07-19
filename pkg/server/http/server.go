package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pliniogsnascimento/pod-scaler-for-tests/pkg/scales"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func StartServer(port string, defaultLogger *logrus.Logger) {
	logger = defaultLogger

	r := gin.Default()
	r.POST("/scaleConfigs", postScaleConfigs)
	r.GET("/scaleConfigs", getScaleConfigs)
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}

func postScaleConfigs(c *gin.Context) {
	var configs scales.ScaleConfigs
	var sleepDuration time.Duration
	var err error
	facade := scales.NewScalesFacade(logger)

	sleepString := c.Request.Header.Get("sleep")
	if sleepDuration, err = time.ParseDuration(sleepString); err != nil {
		sleepDuration = time.Duration(0)
	}

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// clientset, err := facade.GetClientset()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	go facade.UpdateWithConcurrency(configs, &sleepDuration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, map[string]string{"message": "Your request is being processed"})
}

func getScaleConfigs(c *gin.Context) {
	var configs scales.ScaleConfigs
	facade := scales.NewScalesFacade(logger)

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clientset, err := facade.GetClientset()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	currentConfig, err := facade.GetHpaInfo(clientset, configs, logger)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if len(currentConfig) <= 0 {
		c.AbortWithStatus(http.StatusNotFound)
	}

	c.JSON(200, currentConfig)
}
