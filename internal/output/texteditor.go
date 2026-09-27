package output

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

var ErrCanceled = errors.New("canceled")

type TextEditor interface {
	Edit(text string) (string, error)
}

type textEditor struct {
	external string
	shell    executor.Executor
	err      *os.File
	in       *os.File
	out      *os.File
	log      log.Logger
}

func NewTextEditor(shell executor.Executor) *textEditor {
	return &textEditor{
		external: findEditor(runtime.GOOS),
		shell:    shell,
		err:      os.Stderr,
		in:       os.Stdin,
		out:      os.Stdout,
		log:      log.Default(),
	}
}

func (e *textEditor) Edit(text string) (string, error) {
	e.printf("\n%s\n", text)
	p := prompt{in: e.in, out: e.out, err: e.err}
	return p.run(text, []choice{
		{key: "a", label: "[a]ccept", act: func(cur string) (string, bool, error) {
			return cur, true, nil
		}},
		{key: "e", label: "[e]dit", act: func(cur string) (string, bool, error) {
			updated, err := e.edit(cur)
			if err != nil {
				return "", true, fmt.Errorf("failed to edit text: %w", err)
			}
			if updated == "" {
				return "", true, ErrCanceled
			}
			e.printf("\nupdated:\n%s\n", updated)
			return updated, false, nil
		}},
		{key: "c", label: "[c]ancel", act: func(cur string) (string, bool, error) {
			e.printf("%s\n", "canceled.")
			return "", true, ErrCanceled
		}},
		{key: "p", label: "[p]rint", act: func(cur string) (string, bool, error) {
			e.printf("%s\n", cur)
			return "", true, ErrCanceled
		}},
	})
}

func (e *textEditor) edit(input string) (string, error) {
	tmp, err := os.CreateTemp("", "aidy-edittext-*.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil {
			e.log.Error("failed to remove temp file '%s': %v", tmp.Name(), err)
		}
	}()
	if _, err := tmp.WriteString(input); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		e.log.Error("failed to close temp file '%s': %v", tmp.Name(), err)
	}
	parts := strings.Fields(e.external)
	args := append(parts[1:], tmp.Name())
	if _, err := e.shell.RunInteractively(parts[0], args...); err != nil {
		return "", fmt.Errorf("failed to run command '%s %s': %w", e.external, tmp.Name(), err)
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited file '%s': %w", tmp.Name(), err)
	}
	return string(edited), nil
}

func (e *textEditor) printf(format string, args ...any) {
	if _, err := fmt.Fprintf(e.out, format, args...); err != nil {
		panic(err)
	}
}
