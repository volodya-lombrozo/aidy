package output

import (
	"errors"
	"os"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

var ErrCanceled = errors.New("canceled")

type TextReviewer interface {
	Review(text string) (string, error)
}

// textReviewer is reviewer's counterpart for free text: same prompt, same
// options, and the same editor doing the actual editing. Only what the
// buffer means, and what may be done with it, differ.
type textReviewer struct {
	editor *editor
	err    *os.File
	in     *os.File
	out    *os.File
}

func NewTextReviewer(shell executor.Executor) *textReviewer {
	return &textReviewer{
		editor: newEditor(shell),
		err:    os.Stderr,
		in:     os.Stdin,
		out:    os.Stdout,
	}
}

// Review shows text and returns the accepted version. Unlike Print, backing
// out here leaves the caller with nothing to work with, so a canceled
// prompt is reported as ErrCanceled.
func (r *textReviewer) Review(text string) (string, error) {
	printf(r.out, "\n%s\n", text)
	p := prompt{in: r.in, out: r.out, err: r.err}
	reviewed, v, err := p.run(text, []choice{
		acceptChoice(),
		editChoice(r.editor, "text", r.out),
		cancelChoice(r.out),
		printChoice(r.out),
	})
	if err != nil {
		return "", err
	}
	if v == canceled {
		return "", ErrCanceled
	}
	return reviewed, nil
}
