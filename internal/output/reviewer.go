package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

// reviewer shows a generated shell command and lets the user decide what
// becomes of it. It does not edit anything itself - that is what the editor
// it holds is for.
type reviewer struct {
	editor *editor
	shell  executor.Executor
	err    *os.File
	in     *os.File
	out    *os.File
}

func NewReviewer(shell executor.Executor) *reviewer {
	return &reviewer{
		editor: newEditor(shell),
		shell:  shell,
		err:    os.Stderr,
		in:     os.Stdin,
		out:    os.Stdout,
	}
}

// Print shows a generated shell command and offers to run it. Backing out
// is not a failure here - the user simply chose not to run anything - so a
// canceled prompt returns no error.
func (r *reviewer) Print(text Text) error {
	pretty := prettified(text)
	command, err := pretty()
	if err != nil {
		return err
	}
	printf(r.out, "\ngenerated command:\n%s\n", command)
	p := prompt{in: r.in, out: r.out, err: r.err}
	_, _, err = p.run(command, []choice{
		runChoice(r.shell, r.out),
		editChoice(r.editor, "command", r.out),
		generateChoice(pretty, "command", r.out),
		cancelChoice(r.out),
		printChoice(r.out),
	})
	return err
}

func prettified(text Text) Text {
	return func() (string, error) {
		command, err := text()
		if err != nil {
			return "", err
		}
		return prettyCommand(command), nil
	}
}

// runChoice executes the buffer as a shell command. It lives here rather
// than with the general options because only a command buffer can be run.
func runChoice(shell executor.Executor, out io.Writer) choice {
	return choice{key: "r", label: "[r]un", act: func(current string) (string, verdict, error) {
		printf(out, "running...\n")
		parts := cleanQoutes(splitCommand(current))
		_, err := shell.RunInteractively(parts[0], parts[1:]...)
		return current, accepted, err
	}}
}

func cleanQoutes(all []string) []string {
	for i, s := range all {
		all[i] = strings.Trim(s, `"`)
	}
	return all
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
