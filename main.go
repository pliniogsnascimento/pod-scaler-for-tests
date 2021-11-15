/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pliniogsnascimento/pod-scaler-for-tests/pkg/scales"
)

//
// Uncomment to load all auth plugins
// _ "k8s.io/client-go/plugin/pkg/client/auth"
//
// Or uncomment to load specific auth plugins
// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

func main() {
	r := gin.Default()
	r.POST("/scaleConfigs", postScaleConfigs)
	r.GET("/scaleConfigs", getScaleConfigs)
	r.Run("0.0.0.0:8090")
}

func postScaleConfigs(c *gin.Context) {
	logger := log.Default()
	var configs scales.ScaleConfigs

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clientset, err := scales.GetClientset()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	scales.UpdateHpa(clientset, configs, logger)
	c.JSON(200, configs)
}

func getScaleConfigs(c *gin.Context) {
	var configs scales.ScaleConfigs
	logger := log.Default()

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	clientset, err := scales.GetClientset()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	currentConfig, err := scales.GetHpaInfo(clientset, configs, logger)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if len(currentConfig) <= 0 {
		c.AbortWithStatus(http.StatusNotFound)
	}

	c.JSON(200, currentConfig)
}
