package main

import (
	"fmt"
	"os/exec"
	"path"

	tea "github.com/charmbracelet/bubbletea"
)

// sops --config .sops.yaml -e --in-place secretFilePath
func encryptFileWithSops(secretName string) tea.Cmd {
	filePath := path.Join(path.Dir(""), ".sops.yaml")
	fileName := fmt.Sprintf("%s.yaml", secretName)
	fmt.Println(filePath)
	return func() tea.Msg {
		cmd := exec.Command("sops", "--config", filePath, "-e", "--in-place", fileName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
		}
		return out
	}
}
