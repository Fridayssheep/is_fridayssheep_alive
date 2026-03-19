package monitor

import (
	"bytes"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// RunCommand executes a command on the remote server via SSH and returns the output
func RunCommand(client *ssh.Client, cmd string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("SSH client is nil")
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err = session.Run(cmd)
	if err != nil {
		return stdoutBuf.String(), fmt.Errorf("failed to run command: %v, out: %s", err, stdoutBuf.String())
	}

	return stdoutBuf.String(), nil
}
