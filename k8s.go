package main

// import (
// 	"context"
// 	"log"

// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/rest"
// )

// func getClientset() (*kubernetes.Clientset, error) {
// 	config, err := rest.InClusterConfig()

// 	if err != nil {
// 		return nil, err
// 	}

// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return clientset, nil
// }

// func getHpaInfo(clientset *kubernetes.Clientset, scaleConfigs ScaleConfigs, logger *log.Logger) (currentConfig ScaleConfigs) {
// 	for name, _ := range scaleConfigs {
// 		hpa, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(name).Get(context.TODO(), name, metav1.GetOptions{})
// 		if err != nil {
// 			logger.Println(err.Error())
// 			continue
// 		}
// 		// fmt.Fprintf(w, "%s \t%d \t%d \n", hpa.Name, *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas)
// 		currentConfig[hpa.Name] = ScaleConfig{
// 			Min: int(*hpa.Spec.MinReplicas),
// 			Max: int(hpa.Spec.MaxReplicas),
// 		}
// 	}
// 	return currentConfig
// }

// func updateHpa() {

// }
