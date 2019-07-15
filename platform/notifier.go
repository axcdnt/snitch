package platform

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/fishybell/snitch/parser"
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
	Notify(result, pkg string)
}

// DarwinNotifier represents macOS notifier
type DarwinNotifier struct {
}

// Notify notifies desktop notifications on macOS
func (d DarwinNotifier) Notify(result, pkg string) {
	msg := fmt.Sprintf(
		"display notification \"%s\" with title \"%s\" subtitle \"%s\"",
		statusMsg(result),
		"Snitch",
		pkg,
	)
	exec.Command("osascript", "-e", msg).Run()
}

// LinuxNotifier represents Linux notifier
type LinuxNotifier struct {
}

// Notify notifies desktop notifications on Linux
func (l LinuxNotifier) Notify(result, pkg string) {
	msg := fmt.Sprintf("%s: %s", pkg, statusMsg(result))
	err := exec.Command(
		"notify-send", "-a", "Snitch", "-c", "im", "Snitch", msg).Run()
	if err != nil {
		log.Print("Command not found: ", err)
	}
}

func statusMsg(result string) string {
	pass, fail := parser.ParseResult(result)
	return fmt.Sprintf("%d pass, %d fail", pass, fail)
}
