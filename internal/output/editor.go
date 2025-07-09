package output

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/log"
)

type editor struct {
	external string
	shell    executor.Executor
	err      *os.File
	in       *os.File
	out      *os.File
	log      log.Logger
}

func NewEditor(shell executor.Executor) *editor {
	return &editor{
		external: findEditor(runtime.GOOS),
		shell:    shell,
		err:      os.Stderr,
		in:       os.Stdin,
		out:      os.Stdout,
		log:      log.Default(),
	}
}

func (e *editor) Print(command string) error {
	cmd := prettyCommand(command)
	fmt.Printf("\ngenerated command:\n%s\n", cmd)
	for {
		e.printf("%s", "[r]un, [e]dit, [c]ancel, [p]rint? ")
		reader := bufio.NewReader(e.in)
		line, err := reader.ReadString('\n')
		if err != nil {
			e.printfErr("%s: %v\n", "Error reading input", err)
			return err
		}
		choice := strings.ToLower(strings.TrimSpace(line))
		switch choice {
		case "r":
			return e.run(cmd)
		case "e":
			updated := e.edit(cmd)
			if updated == "" {
				return nil
			}
			cmd = updated
			return e.run(cmd)
		case "c":
			e.printf("%s\n", "canceled.")
			return nil
		case "p":
			e.printf("%s\n", cmd)
			return nil
		default:
			e.printfErr("%s\n", "please type r, e, c, or p and press enter.")
		}
	}
}

func (e *editor) run(command string) error {
	e.printf("running...\n")
	parts := cleanQoutes(splitCommand(command))
	_, err := e.shell.RunInteractively(parts[0], parts[1:]...)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func cleanQoutes(all []string) []string {
	for i, s := range all {
		all[i] = strings.Trim(s, `"`)
	}
	return all
}

func (e *editor) printf(format string, args ...any) {
	if _, err := fmt.Fprintf(e.out, format, args...); err != nil {
		panic(err)
	}
}

func (e *editor) printfErr(format string, args ...any) {
	if _, err := fmt.Fprintf(e.err, format, args...); err != nil {
		panic(err)
	}
}

func (e *editor) edit(input string) string {
	tmp, err := os.CreateTemp("", "aidy-editcmd-*.txt")
	if err != nil {
		panic(err)
	}
	e.log.Debug("created temp file '%s' for editing", tmp.Name())
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
	e.log.Debug("temp file '%s' created with content:\n%s", tmp.Name(), input)
	e.log.Debug("using '%s' as editor", e.external)
	editor := e.external
	parts := strings.Fields(editor)
	args := append(parts[1:], tmp.Name())
	_, err = e.shell.RunInteractively(parts[0], args...)
	if err != nil {
		panic(err)
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		panic(err)
	}
	return string(edited)
}

func prettyCommand(command string) string {
	parts := splitCommand(command)
	var res strings.Builder
	for i, p := range parts {
		if strings.HasPrefix(p, "--") {
			res.WriteString("\n")
			res.WriteString("  ")
			res.WriteString(p)
		} else {
			if i != 0 {
				res.WriteString(" ")
			}
			res.WriteString(p)
		}
	}
	return res.String()
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

func splitCommand(input string) []string {
	var args []string
	var current strings.Builder
	quoted := false
	escaped := false
	for _, ch := range strings.TrimSpace(input) {
		switch {
		case escaped:
			current.WriteRune(ch)
			escaped = false
		case ch == '\\':
			escaped = true
		case ch == '"':
			current.WriteRune(ch)
			quoted = !quoted
		case ch == ' ' && !quoted:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case ch == '\n' && !quoted:
			continue
		default:
			current.WriteRune(ch)
		}
	}
	if quoted {
		panic(fmt.Sprintf("unclosed quote in string '%s'", input))
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
