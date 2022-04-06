package tui

import(
"os"
"os/exec"
)

func execCmd(attachStd bool, cmd string, args ...string) error {
  command := exec.Command(cmd, args...)

  if attachStd {
    command.Stdin = os.Stdin
    command.Stdout = os.Stdout
    command.Stderr = os.Stderr
  }
  defer func() {
    command.Stdin = nil
    command.Stdout = nil
    command.Stderr = nil
  }()

  return command.Run()
}
