package main

import (
	"context"
	"fmt"
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
	rootElement.ElementTraversalDisposeNonUpdated()
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
	ctx := ctxElement.Title
	pods, err := ctxElement.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, podItem := range pods.Items {
		ns := podItem.Namespace
		resourceName := podItem.Name
		element := ctxElement.UpsertPod(resourceName, ns)
		if !element.ActionInitialized {
			element.ActionInitialized = true
			element.AddChild("get", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s get pod %s --output=yaml", ns, resourceName))
			element.AddChild("describe", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s describe pod %s", ns, resourceName))
			element.AddChild("logs", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s logs %s", ns, resourceName))
			element.AddChild("logs:follow", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s logs --follow %s", ns, resourceName))
			// element.AddChild("port-forward", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s port-forward %s <port>:<port>", ns, resourceName))
			// element.AddChild("exec:sh", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s exec --stdin --tty %s -- sh", ns, resourceName))
		}
	}
}

func (ctxElement *Element) UpdateDeploymentData() {
	ctx := ctxElement.Title
	deployments, err := ctxElement.Client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, deployItem := range deployments.Items {
		ns := deployItem.Namespace
		resourceName := deployItem.Name
		element := ctxElement.UpsertDeployment(deployItem.Name, deployItem.Namespace)
		if !element.ActionInitialized {
			element.ActionInitialized = true
			element.AddChild("get", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s get deployment %s --output=yaml", ns, resourceName))
			element.AddChild("describe", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s describe deployment %s", ns, resourceName))
		}
	}
}

func (ctxElement *Element) UpdateConfigMapData() {
	ctx := ctxElement.Title
	cms, err := ctxElement.Client.CoreV1().ConfigMaps("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, cmItem := range cms.Items {
		ns := cmItem.Namespace
		resourceName := cmItem.Name
		element := ctxElement.UpsertConfigMap(cmItem.Name, cmItem.Namespace)
		if !element.ActionInitialized {
			element.ActionInitialized = true
			element.AddChild("get", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s get configmap %s --output=yaml", ns, resourceName))
			element.AddChild("describe", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s describe configmap %s", ns, resourceName))
		}
	}
}

func (ctxElement *Element) UpdateServiceData() {
	ctx := ctxElement.Title
	svcs, err := ctxElement.Client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, svcItem := range svcs.Items {
		ns := svcItem.Namespace
		resourceName := svcItem.Name
		element := ctxElement.UpsertService(svcItem.Name, svcItem.Namespace)
		if !element.ActionInitialized {
			element.ActionInitialized = true
			element.AddChild("get", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s get service %s --output=yaml", ns, resourceName))
			element.AddChild("describe", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s describe service %s", ns, resourceName))
			// element.AddChild("port-forward", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s port-forward service/%s <port>:<port>", ns, resourceName))
		}
	}
}

func (ctxElement *Element) UpdateSecretData() {
	ctx := ctxElement.Title
	secrets, err := ctxElement.Client.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		kubeLog.Panic(err)
	}
	for _, secretItem := range secrets.Items {
		ns := secretItem.Namespace
		resourceName := secretItem.Name
		element := ctxElement.UpsertSecret(secretItem.Name, secretItem.Namespace)
		if !element.ActionInitialized {
			element.ActionInitialized = true
			element.AddChild("get", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s get secret %s --output=yaml", ns, resourceName))
			element.AddChild("describe", true).ChannelWaitForCommand(ctx, fmt.Sprintf("kubectl --namespace %s describe secret %s", ns, resourceName))
		}
	}
}
