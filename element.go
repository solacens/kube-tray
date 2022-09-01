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

func (rootElement *Element) UpsertContext(ctx string) *Element {
	if existingCtxElement, ok := rootElement.Children[ctx]; ok {
		existingCtxElement.Updated = true
		return existingCtxElement
	}
	ctxElement := &Element{
		Title:    ctx,
		MenuItem: systray.AddMenuItem(ctx, ctx),
		Children: map[string]*Element{},
		Updated:  true,
	}
	rootElement.Children[ctx] = ctxElement
	// ctxElement.AddChild("Launch console on this context", true).ChannelWaitForShell(ctxElement.Title)
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
	ctxElement.AddChild(ns, false).ChannelWaitForShell(ctxElement.Title, ns)
	return ctxElement
}

func (e *Element) ChannelWaitForManualRefresh(ctx string) {
	go func() {
		for range e.MenuItem.ClickedCh {
			trayLog.Infof("Refresh for %s", ctx)
			rootElement.UpdateContextData(ctx)
		}
	}()
}

func (e *Element) ChannelWaitForShell(ctx string, ns string) {
	go func() {
		for range e.MenuItem.ClickedCh {
			trayLog.Infof("Open shell for %s | %s", ctx, ns)
			OpenTerminal(ctx, ns)
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
