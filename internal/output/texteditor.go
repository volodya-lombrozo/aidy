package output

import (
	"bufio"
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
	reader := bufio.NewReader(e.in)
	for {
		e.printf("%s", "[a]ccept, [e]dit, [c]ancel, [p]rint? ")
		line, err := reader.ReadString('\n')
		if err != nil {
			e.printfErr("%s: %v\n", "Error reading input", err)
			return "", err
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "a":
			return text, nil
		case "e":
			updated, err := e.edit(text)
			if err != nil {
				return "", fmt.Errorf("failed to edit text: %w", err)
			}
			if updated == "" {
				return "", ErrCanceled
			}
			text = updated
			e.printf("\nupdated:\n%s\n", text)
		case "c":
			e.printf("%s\n", "canceled.")
			return "", ErrCanceled
		case "p":
			e.printf("%s\n", text)
			return "", ErrCanceled
		default:
			e.printfErr("%s\n", "please type a, e, c, or p and press enter.")
		}
	}
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

func (e *textEditor) printfErr(format string, args ...any) {
	if _, err := fmt.Fprintf(e.err, format, args...); err != nil {
		panic(err)
	}
}
