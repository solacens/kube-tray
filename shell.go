package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

func OpenTerminal(ctx string, params ...string) {
	shellCommand := viper.GetStringSlice("shell.command")
	cmd := exec.Command(shellCommand[0], append(shellCommand[1:], params...)...)
	tmpEnv := append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", filepath.Join(contextDirectory, ctx)))
	cmd.Env = tmpEnv
	out, err := cmd.CombinedOutput()
	if err != nil {
		trayLog.Fatalf("cmd.Run() failed with %s\n", err)
	}
	trayLog.Debugf("Exec output: %s", out)
}

func RunCommand(ctx string, command string) {
	extraShellRunArgs := viper.GetStringSlice("shell.extraRunArgs")
	OpenTerminal(ctx, append(extraShellRunArgs, command)...)
}
