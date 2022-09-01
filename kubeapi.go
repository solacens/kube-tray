package main

import (
	"context"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (rootElement *Element) UpdateData() {
	kubeLog.Info("Updating Data for all contexts")
	// Mark all for pending deletion
	rootElement.ElementTraversalMarkNonUpdated()
	// Update contexts
	for _, ctx := range existingContext {
		rootElement.UpdateContextData(ctx)
	}
	// Delete missing elements after updates
	rootElement.ElementTraversalDisposeNonUpdated()
}

func GetNamespaces(path string) []v1.Namespace {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		kubeLog.Warning(err)
		return []v1.Namespace{}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		kubeLog.Warning(err)
		return []v1.Namespace{}
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Warning(err)
		return []v1.Namespace{}
	}
	return namespaces.Items
}

func (rootElement *Element) UpdateContextData(ctx string) {
	ctxElement, ok := rootElement.Children[ctx]
	if !ok {
		ctxElement = rootElement.UpsertContext(ctx)
	} else {
		ctxElement.Updated = true
	}

	ctxElement.UpdateNamespaceData()
}

func (ctxElement *Element) UpdateNamespaceData() {
	matches, _ := filepath.Glob(filepath.Join(contextDirectory, ctxElement.Title, "*"))
	for _, match := range matches {
		ns := filepath.Base(match)
		ctxElement.UpsertNamespace(ns)
	}
}
