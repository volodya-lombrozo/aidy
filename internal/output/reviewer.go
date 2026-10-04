package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/volodya-lombrozo/aidy/internal/executor"
)

type Memory interface {
	WithLast(command string)
}

type Replayer interface {
	Replay(command string) error
}

// reviewer shows a generated shell command and lets the user decide what
// becomes of it. It does not edit anything itself - that is what the editor
// it holds is for.
type reviewer struct {
	editor *editor
	shell  executor.Executor
	memory Memory
	err    *os.File
	in     *os.File
	out    *os.File
}

func NewReviewer(shell executor.Executor, memory Memory) *reviewer {
	return &reviewer{
		editor: newEditor(shell),
		shell:  shell,
		memory: memory,
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
	r.memory.WithLast(command)
	return r.review("generated", command, r.remembered(generateChoice(pretty, "command", r.out)))
}

func (r *reviewer) Replay(command string) error {
	return r.review("last", prettyCommand(command))
}

func (r *reviewer) review(origin string, command string, extra ...choice) error {
	printf(r.out, "\n%s command:\n%s\n", origin, command)
	choices := []choice{
		runChoice(r.shell, r.out),
		r.remembered(editChoice(r.editor, "command", r.out)),
	}
	choices = append(choices, extra...)
	choices = append(choices, cancelChoice(r.out), printChoice(r.out))
	p := prompt{in: r.in, out: r.out, err: r.err}
	_, _, err := p.run(command, choices)
	return err
}

func (r *reviewer) remembered(c choice) choice {
	act := c.act
	c.act = func(current string) (string, verdict, error) {
		next, v, err := act(current)
		if v == again {
			r.memory.WithLast(next)
		}
		return next, v, err
	}
	return c
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
