package output

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// verdict is what an action decided about the loop it runs in. Keeping
// "the user gave up" separate from "something went wrong" is what lets the
// same options serve prompts that disagree about whether a cancel is an
// error: the choice reports canceled, the caller decides what it means.
type verdict int

const (
	// again re-prompts with the buffer the action handed back.
	again verdict = iota
	// accepted stops the loop and keeps the buffer.
	accepted
	// canceled stops the loop and discards the buffer.
	canceled
)

// choice is one action offered at a prompt loop. key is the single
// lowercase letter that selects it, label is how it's rendered in the
// hint line (e.g. "[r]un"), and act performs the action against the
// current buffer, returning the buffer to keep (possibly unchanged) and
// what should happen to the loop.
type choice struct {
	key   string
	label string
	act   func(current string) (next string, v verdict, err error)
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

func (p *prompt) run(initial string, choices []choice) (string, verdict, error) {
	current := initial
	reader := bufio.NewReader(p.in)
	for {
		p.printf("%s? ", hint(choices))
		line, err := reader.ReadString('\n')
		if err != nil {
			p.printfErr("%s: %v\n", "Error reading input", err)
			return "", canceled, err
		}
		key := strings.ToLower(strings.TrimSpace(line))
		matched := false
		for _, c := range choices {
			if c.key == key {
				matched = true
				next, v, err := c.act(current)
				current = next
				if v != again {
					return current, v, err
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
	printf(p.out, format, args...)
}

func (p *prompt) printfErr(format string, args ...any) {
	printf(p.err, format, args...)
}

// printf is the package's one way of writing to a stream. Every writer here
// is a pipe or a standard stream, so a failure means something is wrong that
// this package cannot do anything sensible about.
func printf(out io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(out, format, args...); err != nil {
		panic(err)
	}
}
