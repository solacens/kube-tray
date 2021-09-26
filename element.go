package main

import (
	"github.com/getlantern/systray"
	"k8s.io/client-go/kubernetes"
)

type Element struct {
	Title             string
	MenuItem          *systray.MenuItem
	Children          map[string]*Element
	Client            *kubernetes.Clientset
	ActionInitialized bool
	Updated           bool
	Locked            bool
}

func NewRoot() *Element {
	return &Element{
		MenuItem: nil,
		Children: map[string]*Element{},
		Updated:  true,
		Locked:   true,
	}
}

func (e *Element) AddChild(title string, locked bool) *Element {
	element := &Element{
		Title:    title,
		MenuItem: e.MenuItem.AddSubMenuItem(title, title),
		Children: map[string]*Element{},
		Client:   e.Client,
		Updated:  true,
		Locked:   locked,
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
		Children: map[string]*Element{},
		Client:   client,
		Updated:  true,
	}
	rootElement.Children[ctx] = ctxElement
	ctxElement.AddChild("Launch console on this context", true).ChannelWaitForShell(ctxElement.Title)
	ctxElement.AddChild("Refresh", true).ChannelWaitForManualRefresh(ctxElement.Title)
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
		return existingNsElement.Children["Pod"].AddChild(pod, false)
	}
	return nil
}

func (ctxElement *Element) UpsertDeployment(deploy string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Deployment"].Children[deploy]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		return existingNsElement.Children["Deployment"].AddChild(deploy, false)
	}
	return nil
}

func (ctxElement *Element) UpsertConfigMap(cm string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["ConfigMap"].Children[cm]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		return existingNsElement.Children["ConfigMap"].AddChild(cm, false)
	}
	return nil
}

func (ctxElement *Element) UpsertService(svc string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Service"].Children[svc]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		return existingNsElement.Children["Service"].AddChild(svc, false)
	}
	return nil
}

func (ctxElement *Element) UpsertSecret(secret string, ns string) *Element {
	if existingNsElement, ok := ctxElement.Children[ns]; ok {
		if existingElement, ok2 := existingNsElement.Children["Secret"].Children[secret]; ok2 {
			existingElement.Updated = true
			return existingElement
		}
		return existingNsElement.Children["Secret"].AddChild(secret, false)
	}
	return nil
}

func (e *Element) ChannelWaitForManualRefresh(ctx string) {
	go func() {
		for range e.MenuItem.ClickedCh {
			trayLog.Infof("Refresh for %s", ctx)
			rootElement.UpdateContextData(ctx)
		}
	}()
}

func (e *Element) ChannelWaitForShell(ctx string) {
	go func() {
		for range e.MenuItem.ClickedCh {
			trayLog.Infof("Open shell for %s", ctx)
			OpenTerminal(ctx)
		}
	}()
}

func (element *Element) ChannelWaitForCommand(ctx string, cmd string) {
	go func() {
		for range element.MenuItem.ClickedCh {
			trayLog.Infof("Run command in %s: %s", ctx, cmd)
			RunCommand(ctx, cmd)
		}
	}()
}

func (e *Element) ElementTraversalMarkNonUpdated() {
	if !e.Locked {
		e.Updated = false
	}
	for _, childElement := range e.Children {
		childElement.ElementTraversalMarkNonUpdated()
	}
}

func (e *Element) ElementTraversalDisposeNonUpdated() {
	if !e.Updated {
		e.MenuItem.Hide()
		// Force hide and close channel on gone resources
		if e.ActionInitialized {
			for _, childElement := range e.Children {
				childElement.MenuItem.Hide()
				close(childElement.MenuItem.ClickedCh)
			}
		}
	}
	for _, childElement := range e.Children {
		childElement.ElementTraversalDisposeNonUpdated()
	}
}
