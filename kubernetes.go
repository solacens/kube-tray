package main

import (
	"context"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (rootElement *Element) UpdateData() {
	// Mark all for pending deletion
	rootElement.ElementTraversalMarkNonUpdated()
	// Update contexts
	for _, ctx := range existingContext {
		rootElement.UpdateContextData(ctx)
	}
	// Delete missing elements after updates
	rootElement.ElementTraversalHideNonUpdated()
}

func (rootElement *Element) UpdateContextData(ctx string) {
	ctxElement, ok := rootElement.Children[ctx]
	if !ok {
		config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(contextDirectory, ctx))
		if err != nil {
			kubeLog.Panic(err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			kubeLog.Panic(err)
		}
		ctxElement = rootElement.UpsertContext(ctx, clientset)
	} else {
		ctxElement.Updated = true
	}

	ctxElement.UpdateNamespaceData()
}

func (ctxElement *Element) UpdateNamespaceData() {
	namespaces, err := ctxElement.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, nsItem := range namespaces.Items {
		ns := nsItem.Name
		ctxElement.UpsertNamespace(ns)
	}
	ctxElement.UpdatePodData()
	ctxElement.UpdateDeploymentData()
	ctxElement.UpdateConfigMapData()
	ctxElement.UpdateServiceData()
	ctxElement.UpdateSecretData()
}

func (ctxElement *Element) UpdatePodData() {
	pods, err := ctxElement.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, podItem := range pods.Items {
		ctxElement.UpsertPod(podItem.Name, podItem.Namespace)
	}
}

func (ctxElement *Element) UpdateDeploymentData() {
	deployments, err := ctxElement.Client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, deployItem := range deployments.Items {
		ctxElement.UpsertDeployment(deployItem.Name, deployItem.Namespace)
	}
}

func (ctxElement *Element) UpdateConfigMapData() {
	cms, err := ctxElement.Client.CoreV1().ConfigMaps("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, cmItem := range cms.Items {
		ctxElement.UpsertConfigMap(cmItem.Name, cmItem.Namespace)
	}
}

func (ctxElement *Element) UpdateServiceData() {
	svcs, err := ctxElement.Client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, svcItem := range svcs.Items {
		ctxElement.UpsertService(svcItem.Name, svcItem.Namespace)
	}
}

func (ctxElement *Element) UpdateSecretData() {
	secrets, err := ctxElement.Client.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, secretItem := range secrets.Items {
		ctxElement.UpsertSecret(secretItem.Name, secretItem.Namespace)
	}
}
