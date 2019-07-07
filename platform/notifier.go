package platform

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/axcdnt/snitch/parser"
)

// NewNotifier creates a notifier
func NewNotifier() Notifier {
	switch runtime.GOOS {
	case "darwin":
		return DarwinNotifier{}
	case "linux":
		return LinuxNotifier{}
	}

	return nil
}

// Notifier represents a platform notifier
type Notifier interface {
	Notify(status, pkg string)
}

// DarwinNotifier represents macOS notifier
type DarwinNotifier struct {
}

// Notify notifies desktop notifications on macOS
func (d DarwinNotifier) Notify(output, pkg string) {
	msg := fmt.Sprintf(
		"display notification \"%s\" with title \"%s\" subtitle \"%s\"",
		status(output),
		"Snitch",
		pkg,
	)
	exec.Command("osascript", "-e", msg).Run()
}

// LinuxNotifier represents Linux notifier
type LinuxNotifier struct {
}

// Notify notifies desktop notifications on Linux
func (l LinuxNotifier) Notify(output, pkg string) {
	msg := fmt.Sprintf("%s: %s", pkg, status(output))
	err := exec.Command(
		"notify-send", "-a", "Snitch", "-c", "im", "Snitch", msg).Run()
	if err != nil {
		log.Print("Command not found: ", err)
	}
}

func status(output string) string {
	pass, fail := parser.ParseOutput(output)
	return fmt.Sprintf("%d pass, %d fail", pass, fail)
}
