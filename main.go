package main

import (
	"io"
	"os"
	"path/filepath"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"

	"github.com/getlantern/systray"
	"github.com/solacens/kube-tray/icon"

	"k8s.io/client-go/util/homedir"
)

var (
	contextDirectory string
	existingContext  []string

	rootElement *Element

	trayLog *log.Entry
	kubeLog *log.Entry
)

func init() {
	// File paths
	home := homedir.HomeDir()
	logFilePath := filepath.Join(home, ".kube-tray", "log.")
	// appConfigPath = filepath.Join(home, ".kube-tray", "config")

	// Logger setting
	r, _ := rotatelogs.New(logFilePath + "%Y%m%d")
	mw := io.MultiWriter(os.Stdout, r)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(mw)
	log.SetLevel(log.DebugLevel)
	trayLog = log.WithFields(log.Fields{
		"type": "tray",
	})
	kubeLog = log.WithFields(log.Fields{
		"type": "kube",
	})
}

func main() {
	LoadKubeconfig(false)

	systray.Run(onTrayReady, func() {})
}

func onTrayReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("K8S Tray")
	systray.SetTooltip("Kubernetes Tray")

	//////////////////////////////////
	quitMenuItem := systray.AddMenuItem("Quit", "Quit")

	//////////////////////////////////
	systray.AddSeparator()

	//////////////////////////////////
	reloadMenuItem := systray.AddMenuItem("Reload Kubeconfig", "Reload Kubeconfig")
	reloadMenuItemFunc := func() {
		trayLog.Info("Reloading config")
		LoadKubeconfig(true)
		go rootElement.UpdateData()
		trayLog.Info("Reloaded")
	}

	//////////////////////////////////
	systray.AddSeparator()

	//////////////////////////////////
	rootElement = NewRoot()
	go rootElement.UpdateData()

	//////////////////////////////////
	trayLog.Info("Ready")

	//////////////////////////////////
	for {
		select {
		case <-quitMenuItem.ClickedCh:
			trayLog.Info("Quit")
			systray.Quit()
		case <-reloadMenuItem.ClickedCh:
			reloadMenuItemFunc()
		}
	}
}
