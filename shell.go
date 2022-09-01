package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

func OpenTerminal(ctx string, ns string) {
	shellCommand := viper.GetStringSlice("shell.command")
	kubeconfigPath := filepath.Join(contextDirectory, ctx, ns)
	cmd := exec.Command(shellCommand[0], shellCommand[1:]...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		trayLog.Fatalf("cmd.Run() failed with %s\n", err)
	}
	trayLog.Debugf("Exec output: %s", out)
}
