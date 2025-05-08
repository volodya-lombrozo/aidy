package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/volodya-lombrozo/aidy/executor"
)

type Output interface {
	Print(command string)
}

type editor struct {
	external string
	shell    executor.Executor
	err      *os.File
	in       *os.File
	out      *os.File
}

type printer struct {
}

type Mock struct {
	captured []string
}

func NewEditor(shell executor.Executor) *editor {
	return &editor{
		external: findEditor(runtime.GOOS),
		shell:    shell,
		err:      os.Stderr,
		in:       os.Stdin,
		out:      os.Stdout,
	}
}

func NewPrinter() Output {
	return &printer{}
}

func NewMock() *Mock {
	return &Mock{}
}

func (e *editor) Print(command string) {
	cmd := prettyCommand(command)
	fmt.Printf("\ngenerated command:\n%s\n", cmd)
	for {
		e.printf("%s", "[r]un, [e]dit, [c]ancel, [p]rint? ")
		reader := bufio.NewReader(e.in)
		line, err := reader.ReadString('\n')
		if err != nil {
			e.printfErr("%s: %v\n", "Error reading input", err)
			return
		}
		choice := strings.ToLower(strings.TrimSpace(line))
		switch choice {
		case "r":
			e.run(cmd)
			return
		case "e":
			updated := e.edit(cmd)
			if updated == "" {
				return
			}
			cmd = updated
			e.run(cmd)
			return
		case "c":
			e.printf("%s\n", "canceled.")
			return
		case "p":
			e.printf("%s\n", cmd)
			return
		default:
			e.printfErr("%s\n", "please type r, e, c, or p and press enter.")
		}
	}
}

func (e *editor) run(command string) {
	e.printf("running '%s' command\n", command)
	parts := cleanQoutes(splitCommand(command))
	e.printf("command parts '%v'\n", parts)
	_, err := e.shell.RunInteractively(parts[0], parts[1:]...)
	if err != nil {
		panic(err)
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
	log.Printf("created the temp file '%s' for editing\n", tmp.Name())
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil {
			log.Printf("failed to remove temp file: %v", err)
		}
	}()
	if _, err := tmp.WriteString(input); err != nil {
		panic(err)
	}
	if err := tmp.Close(); err != nil {
		log.Printf("failed to close temp file: %v", err)
	}
	log.Printf("use '%s' program for editing\n", e.external)
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

func (p *printer) Print(command string) {
	fmt.Println(command)
}

func (m *Mock) Print(command string) {
	m.captured = append(m.captured, command)
}

func (m *Mock) Last() string {
	size := len(m.captured)
	if size < 1 {
		panic("we weren't able to capture anything")
	}
	return m.captured[size-1]
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
