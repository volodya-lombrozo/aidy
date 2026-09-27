package output

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// choice is one action offered at a prompt loop. key is the single
// lowercase letter that selects it, label is how it's rendered in the
// hint line (e.g. "[r]un"), and act performs the action against the
// current buffer, returning the buffer to keep (possibly unchanged) and
// whether the loop should stop.
type choice struct {
	key   string
	label string
	act   func(current string) (next string, done bool, err error)
}

// prompt drives a "type a letter, run its action, repeat" loop shared by
// every run/edit/cancel-style confirmation in this package. It only knows
// how to render choices, read a line, and dispatch - it has no idea what
// the choices actually do.
type prompt struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (p *prompt) run(initial string, choices []choice) (string, error) {
	current := initial
	reader := bufio.NewReader(p.in)
	for {
		p.printf("%s? ", hint(choices))
		line, err := reader.ReadString('\n')
		if err != nil {
			p.printfErr("%s: %v\n", "Error reading input", err)
			return "", err
		}
		key := strings.ToLower(strings.TrimSpace(line))
		matched := false
		for _, c := range choices {
			if c.key == key {
				matched = true
				next, done, err := c.act(current)
				current = next
				if done {
					return current, err
				}
				break
			}
		}
		if !matched {
			p.printfErr("please type %s and press enter.\n", keys(choices))
		}
	}
}

func hint(choices []choice) string {
	labels := make([]string, len(choices))
	for i, c := range choices {
		labels[i] = c.label
	}
	return strings.Join(labels, ", ")
}

// keys renders the choice keys as "r, e, c, or p", matching how the
// invalid-option message has always read.
func keys(choices []choice) string {
	all := make([]string, len(choices))
	for i, c := range choices {
		all[i] = c.key
	}
	if len(all) < 2 {
		return strings.Join(all, "")
	}
	return strings.Join(all[:len(all)-1], ", ") + ", or " + all[len(all)-1]
}

func (p *prompt) printf(format string, args ...any) {
	if _, err := fmt.Fprintf(p.out, format, args...); err != nil {
		panic(err)
	}
}

func (p *prompt) printfErr(format string, args ...any) {
	if _, err := fmt.Fprintf(p.err, format, args...); err != nil {
		panic(err)
	}
}
