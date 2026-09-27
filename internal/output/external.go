package output

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

// external is the user's $EDITOR. It is the single place that knows the
// temp-file dance: write the buffer out, hand the file to the editor, read
// back whatever was saved, clean up.
type external struct {
	cmd   string
	shell executor.Executor
	log   log.Logger
}

func newExternal(shell executor.Executor) *external {
	return &external{
		cmd:   findEditor(runtime.GOOS),
		shell: shell,
		log:   log.Default(),
	}
}

// open edits text in the external editor and returns what the user saved.
// noun names what is being edited ("command", "text") and shows up in the
// temp file name so a leftover file says where it came from.
func (e *external) open(noun string, text string) (string, error) {
	tmp, err := os.CreateTemp("", fmt.Sprintf("aidy-edit%s-*.txt", noun))
	if err != nil {
		panic(err)
	}
	e.log.Debug("created temp file '%s' for editing", tmp.Name())
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil {
			e.log.Error("failed to remove temp file '%s': %v", tmp.Name(), err)
		}
	}()
	if _, err := tmp.WriteString(text); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		e.log.Error("failed to close temp file '%s': %v", tmp.Name(), err)
	}
	e.log.Debug("temp file '%s' created with content:\n%s", tmp.Name(), text)
	e.log.Debug("using '%s' as editor", e.cmd)
	parts := strings.Fields(e.cmd)
	args := append(parts[1:], tmp.Name())
	if _, err := e.shell.RunInteractively(parts[0], args...); err != nil {
		return "", fmt.Errorf("failed to run command '%s %s': %w", e.cmd, tmp.Name(), err)
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file '%s': %w", tmp.Name(), err)
	}
	return string(edited), nil
}

func findEditor(runtime string) string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	switch runtime {
	case "windows":
		return "notepad"
	default:
		return "vi"
	}
}
