package main

import (
	"context"
	"strings"

	"github.com/posener/complete"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CompletionPredictors() map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"file":              complete.PredictFiles("*"),
		"directory":         complete.PredictDirs("*"),
		"file_or_directory": complete.PredictOr(complete.PredictFiles("*"), complete.PredictDirs("*")),
		"namespace":         namespacePredictor(), 
		"context":           contextPredictor(),
	}
}

func contextPredictor() complete.PredictFunc {
	return func(a complete.Args) (prediction []string) {
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		
		kubeConfig, err := clientConfig.RawConfig()
		if err != nil {
			return
		}

		var predictions []string
		for name := range kubeConfig.Contexts {
			if strings.HasPrefix(name, a.Last) {
				predictions = append(predictions, name)
			}
		}
		return predictions
	}
}

func namespacePredictor() complete.PredictFunc {
	return func(a complete.Args) (prediction []string) {
		client, err := k8sClient(); 
		if err != nil {
			return
		}

		namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return
		}

		var predictions []string
		for _, ns := range namespaceList.Items {
			if strings.HasPrefix(ns.GetName(), a.Last) {
				predictions = append(predictions, ns.GetName())
			}
		}
		return predictions
	}
}

func k8sClient() (*kubernetes.Clientset, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	kubeConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}


	return kubernetes.NewForConfig(kubeConfig)
}