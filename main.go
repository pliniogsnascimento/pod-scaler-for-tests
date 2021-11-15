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

	"context"
	"github.com/gin-gonic/gin"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

type ScaleConfigs map[string]ScaleConfig

type ScaleConfig struct {
	Min         int  `json:"min"`
	Max         int  `json:"max"`
	HpaOperator bool `json:"hpaOperator"`
}

func main() {
	r := gin.Default()
	r.POST("/scaleConfigs", postScaleConfigs)
	r.GET("/scaleConfigs", getScaleConfigs)
	r.Run("0.0.0.0:8090")
}

func postScaleConfigs(c *gin.Context) {
	// logger := log.Default()
	var configs ScaleConfigs

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, configs)
}

func getScaleConfigs(c *gin.Context) {
	var configs ScaleConfigs
	logger := log.Default()

	if err := c.ShouldBindJSON(&configs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	clientset, err := getClientset()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	currentConfig := getHpaInfo(clientset, configs, logger)
	if len(currentConfig) <= 0 {
		c.AbortWithStatus(http.StatusNotFound)
	}

	c.JSON(200, currentConfig)
}

// TODO: Remove from here after resolving bugs

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()

	if err != nil {

		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func getHpaInfo(clientset *kubernetes.Clientset, scaleConfigs ScaleConfigs, logger *log.Logger) (currentConfig ScaleConfigs) {
	for name, _ := range scaleConfigs {
		hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(name).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			logger.Println(err.Error())
			continue
		}
		// fmt.Fprintf(w, "%s \t%d \t%d \n", hpa.Name, *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas)
		currentConfig[hpa.Name] = ScaleConfig{
			Min: int(*hpa.Spec.MinReplicas),
			Max: int(hpa.Spec.MaxReplicas),
		}
	}
	return currentConfig
}

func updateHpa() {

}
