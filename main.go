package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/getlantern/systray"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"github.com/solacens/kube-tray/icon"
	"github.com/spf13/viper"
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
	// Home directory
	home := homedir.HomeDir()

	// Logger setting
	logFilePath := filepath.Join(home, ".kube-tray", "log.")
	r, _ := rotatelogs.New(logFilePath + "%Y%m%d")
	mw := io.MultiWriter(os.Stdout, r)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(mw)
	log.SetLevel(log.InfoLevel)
	trayLog = log.WithFields(log.Fields{
		"type": "tray",
	})
	kubeLog = log.WithFields(log.Fields{
		"type": "kube",
	})

	// Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(home, ".kube-tray"))
	err := viper.ReadInConfig()
	if err != nil {
		// Default terminal command
		if runtime.GOOS == "windows" {
			// cmd /c wt -w 0 nt
			viper.Set("shell.command", []string{"cmd", "/c", "wt", "-w", "0", "nt"})
			// cmd /c wt -w 0 nt pwsh -noe -c <command>
			viper.Set("shell.extraRunArgs", []string{"pwsh", "-noe", "-c"})
		} else {
			viper.Set("shell.command", []string{"bash"}) // TODO: Darwin & Linux
		}
		// Auto refresh
		viper.Set("auto-refresh.enabled", false)
		viper.Set("auto-refresh.interval", 10)
		// WriteConfigAs requried for first time creation
		viper.WriteConfigAs(filepath.Join(home, ".kube-tray", "config.yaml"))
	}
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
