package main

import (
	"context"
	"encoding/json"
	"strings"

	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
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
		"k8s_resource":      kubernetesResourcePredictor(),
	}
}

func kubernetesResourcePredictor() complete.PredictFunc {
	return func(a complete.Args) (prediction []string) {
		client, err := k8sClient()
		if err != nil {
			return
		}
		resources, err := client.RESTClient().
			Get().
			AbsPath("/apis/apiextensions.k8s.io/v1/CustomResourceDefinition").
			Resource(v1.CompositionKind).
			DoRaw(context.TODO())

		if err != nil {
			return
		}

		rl := metav1.APIResourceList{}
		if err := json.Unmarshal(resources, &rl); err != nil {
			return
		}

		var predictions []string
		for _, res := range rl.APIResources {
			if strings.HasPrefix(res.Name, a.Last) {
				predictions = append(predictions, res.Name)
			}
		}
		return predictions
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
		client, err := k8sClient()
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
