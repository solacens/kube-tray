package main

import (
	"github.com/getlantern/systray"

	"k8s.io/client-go/kubernetes"
)

type Element struct {
	Title      string
	MenuItem   *systray.MenuItem
	Client     *kubernetes.Clientset
	Children   map[string]*Element
	Updated    bool
	FixUpdated bool
}

func NewRoot() *Element {
	return &Element{
		MenuItem: nil,
		Children: map[string]*Element{},
		Updated:  true,
	}
}

func (e *Element) AddChild(title string, fixUpdated bool) *Element {
	element := &Element{
		Title:      title,
		MenuItem:   e.MenuItem.AddSubMenuItem(title, title),
		Client:     e.Client,
		Children:   map[string]*Element{},
		Updated:    true,
		FixUpdated: fixUpdated,
	}
	e.Children[title] = element
	return element
}

func (rootElement *Element) UpsertContext(ctx string, client *kubernetes.Clientset) *Element {
	if existingCtxElement, ok := rootElement.Children[ctx]; ok {
		existingCtxElement.Updated = true
		return existingCtxElement
	}
	ctxElement := &Element{
		Title:    ctx,
		MenuItem: systray.AddMenuItem(ctx, ctx),
		Client:   client,
		Children: map[string]*Element{},
		Updated:  true,
	}
	rootElement.Children[ctx] = ctxElement
	ctxElement.AddChild("Launch console on this context", true)
	ctxElement.AddChild("Refresh", true)
	seperator := ctxElement.MenuItem.AddSubMenuItem("", "")
	seperator.Disable()
	return ctxElement
}

func (ctxElement *Element) UpsertNamespace(ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		existingNsElement.Updated = true
		return existingNsElement
	}
	nsElement := ctxElement.AddChild(ns, false)

	// Pod
	nsElement.AddChild("Pod", true)
	// Deployment
	nsElement.AddChild("Deployment", true)
	// Configmap
	nsElement.AddChild("ConfigMap", true)
	// Service
	nsElement.AddChild("Service", true)
	// Secret
	nsElement.AddChild("Secret", true)

	return nsElement
}

func (ctxElement *Element) UpsertPod(pod string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Pod"].Children[pod]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		element := existingNsElement.Children["Pod"].AddChild(pod, false)
		element.AddChild("get", true)
		element.AddChild("describe", true)
		element.AddChild("logs", true)
		element.AddChild("logs:tail", true)
		element.AddChild("port-forward", true)
		element.AddChild("exec:sh", true)
		return element
	}
	return nil
}

func (ctxElement *Element) UpsertDeployment(deploy string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Deployment"].Children[deploy]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		element := existingNsElement.Children["Deployment"].AddChild(deploy, false)
		element.AddChild("get", true)
		element.AddChild("describe", true)
		return element
	}
	return nil
}

func (ctxElement *Element) UpsertConfigMap(cm string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["ConfigMap"].Children[cm]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		element := existingNsElement.Children["ConfigMap"].AddChild(cm, false)
		element.AddChild("get", true)
		element.AddChild("describe", true)
		return element
	}
	return nil
}

func (ctxElement *Element) UpsertService(svc string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Service"].Children[svc]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		element := existingNsElement.Children["Service"].AddChild(svc, false)
		element.AddChild("get", true)
		element.AddChild("describe", true)
		element.AddChild("port-forward", true)
		return element
	}
	return nil
}

func (ctxElement *Element) UpsertSecret(secret string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Secret"].Children[secret]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		element := existingNsElement.Children["Secret"].AddChild(secret, false)
		element.AddChild("get", true)
		element.AddChild("describe", true)
		return element
	}
	return nil
}

func (e *Element) ElementTraversalMarkNonUpdated() {
	if !e.FixUpdated {
		e.Updated = false
	}
	for _, childElement := range e.Children {
		childElement.ElementTraversalMarkNonUpdated()
	}
}

func (e *Element) ElementTraversalHideNonUpdated() {
	if e.MenuItem != nil && !e.Updated {
		e.MenuItem.Hide()
	}
	for _, childElement := range e.Children {
		childElement.ElementTraversalHideNonUpdated()
	}
}
