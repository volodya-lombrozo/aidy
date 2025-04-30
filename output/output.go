package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Output interface {
	Print(command string)
}

type editor struct {
	external string
}

type printer struct {
}

type Mock struct {
	captured []string
}

func NewEditor() Output {
	return &editor{external: external()}
}

func NewPrinter() Output {
	return &printer{}
}

func NewMock() *Mock {
	return &Mock{}
}

func (e *editor) Print(command string) {
	fmt.Printf("generated command: %s\n", command)
	for {
		fmt.Fprint(os.Stderr, "[r]un, [e]dit, [c]ancel, [p]rint? ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			return
		}
		choice := strings.ToLower(strings.TrimSpace(line))
		switch choice {
		case "r":
			run(command)
			return
		case "e":
			updated := e.edit(command)
			if updated == "" {
				return
			}
			command = updated
			run(command)
			return
		case "c":
			fmt.Println("Canceled.")
			return
		case "p":
			fmt.Println(command)
			return
		default:
			fmt.Fprintln(os.Stderr, "please type r, e, c, or p and press enter.")
		}
	}
}

func run(command string) {
	fmt.Printf("Running '%s' command\n", command)
	parts := splitCommand(command)
	fmt.Printf("command parts '%v'\n", parts)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
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
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		panic(err)
	}
	return string(edited)
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
		panic("We weren't able to capture anything")
	}
	return m.captured[size-1]
}

func external() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	switch runtime.GOOS {
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
			quoted = !quoted
		case ch == ' ' && !quoted:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
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
