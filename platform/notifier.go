package platform

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
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
	Notify(status, dir string)
}

// DarwinNotifier represents macOS notifier
type DarwinNotifier struct {
}

// Notify notifies desktop notifications on macOS
func (d DarwinNotifier) Notify(output, dir string) {
	pass, fail := parser.ParseOutput(output)
	status := fmt.Sprintf("%d success %d fail", pass, fail)
	subtitle := filepath.Base(dir)
	msg := fmt.Sprintf(
		"display notification \"%s\" with title \"%s\" subtitle \"%s\"", status, "Snitch", subtitle)
	exec.Command("osascript", "-e", msg).Run()
}

// LinuxNotifier represents Linux notifier
type LinuxNotifier struct {
}

// Notify notifies desktop notifications on Linux
func (l LinuxNotifier) Notify(output, dir string) {
	pass, fail := parser.ParseOutput(output)
	status := fmt.Sprintf("%d success %d fail", pass, fail)
	msg := fmt.Sprintf("'%s %s'", dir, status)
	err := exec.Command(
		"notify-send", "-a", "Snitch", "-c", "im", "Snitch", msg).Run()
	if err != nil {
		log.Print("Command not found: ", err)
	}
}
