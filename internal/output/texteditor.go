package output

import (
	"errors"
	"os"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

var ErrCanceled = errors.New("canceled")

type TextEditor interface {
	Edit(text string) (string, error)
}

// textEditor reviews free text the same way reviewer reviews a command, and
// leans on the same editor to do the actual editing.
type textEditor struct {
	editor *editor
	err    *os.File
	in     *os.File
	out    *os.File
}

func NewTextEditor(shell executor.Executor) *textEditor {
	return &textEditor{
		editor: newEditor(shell),
		err:    os.Stderr,
		in:     os.Stdin,
		out:    os.Stdout,
	}
}

// Edit shows text for review and returns the accepted version. Unlike
// Print, backing out here leaves the caller with nothing to work with, so a
// canceled prompt is reported as ErrCanceled.
func (e *textEditor) Edit(text string) (string, error) {
	printf(e.out, "\n%s\n", text)
	p := prompt{in: e.in, out: e.out, err: e.err}
	reviewed, v, err := p.run(text, []choice{
		acceptChoice(),
		editChoice(e.editor, "text", e.out),
		cancelChoice(e.out),
		printChoice(e.out),
	})
	if err != nil {
		return "", err
	}
	if v == canceled {
		return "", ErrCanceled
	}
	return reviewed, nil
}
