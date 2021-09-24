package main

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func LoadKubeconfig(clean bool) {
	home := homedir.HomeDir()
	contextDirectory = filepath.Join(home, ".kube-tray", "contexts")
	if _, err := os.Stat(contextDirectory); !os.IsNotExist(err) {
		if clean {
			os.RemoveAll(contextDirectory)
		} else {
			matches, _ := filepath.Glob(filepath.Join(contextDirectory, "*"))
			for _, match := range matches {
				ctx := filepath.Base(match)
				existingContext = append(existingContext, ctx)
				kubeLog.Info(fmt.Sprintf("Loaded kubeconfig [%s]", ctx))
			}
			return
		}
	} else {
		os.MkdirAll(contextDirectory, 0755)
	}

	// Load default kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.RawConfig()
	if err != nil {
		kubeLog.Panic(err)
	}

	// Create seperate kubeconfig per context
	for ctx := range config.Contexts {
		user := config.Contexts[ctx].AuthInfo
		newConfig := *clientcmdapi.NewConfig()
		newConfig.Contexts[ctx] = config.Contexts[ctx].DeepCopy()
		newConfig.AuthInfos[user] = config.AuthInfos[user].DeepCopy()
		newConfig.Clusters[ctx] = config.Clusters[ctx].DeepCopy()
		newConfig.CurrentContext = ctx

		existingContext = append(existingContext, ctx)
		kubeLog.Info(fmt.Sprintf("(Re)Created kubeconfig [%s]", ctx))

		clientcmd.WriteToFile(newConfig, filepath.Join(contextDirectory, ctx))
	}
}
