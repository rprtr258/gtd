package cmds

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

const GTD_DIR = "/home/rprtr258/GTD/"

// Run executable with arguments
func Run(ctx context.Context, executable string, args ...string) error {
	executable, err := exec.LookPath(executable)
	if err != nil {
		return err
	}

	if _, err = os.StartProcess(
		executable,
		append([]string{executable}, args...),
		&os.ProcAttr{
			Dir:   ".",
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, nil, nil},
			Sys:   &syscall.SysProcAttr{},
		},
	); err != nil {
		return err
	}

	return nil
}

// Open file
func Open(ctx context.Context, open_what string) error {
	return Run(ctx, "/usr/bin/open", open_what)
}
