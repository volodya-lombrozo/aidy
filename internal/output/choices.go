package output

import (
	"fmt"
	"io"
)

// This is the library of prompt options. Each constructor returns one
// self-contained choice, so a prompt is assembled by listing the options it
// wants instead of by extending a switch. An option that is useful in only
// one place still belongs here if a second place might ever want it.

// acceptChoice keeps the buffer as it stands and finishes.
func acceptChoice() choice {
	return choice{key: "a", label: "[a]ccept", act: func(current string) (string, verdict, error) {
		return current, accepted, nil
	}}
}

// editChoice hands the buffer to the external editor and comes back for
// another round, so edits can be stacked. An editor that saved nothing is
// treated as a cancel. noun names what is being edited ("command", "text")
// and appears in the messages this option prints.
func editChoice(ext *external, noun string, out io.Writer) choice {
	return choice{key: "e", label: "[e]dit", act: func(current string) (string, verdict, error) {
		updated, err := ext.open(noun, current)
		if err != nil {
			return "", canceled, fmt.Errorf("failed to edit %s: %w", noun, err)
		}
		if updated == "" {
			return "", canceled, nil
		}
		printf(out, "\nupdated %s:\n%s\n", noun, updated)
		return updated, again, nil
	}}
}

// cancelChoice throws the buffer away.
func cancelChoice(out io.Writer) choice {
	return choice{key: "c", label: "[c]ancel", act: func(current string) (string, verdict, error) {
		printf(out, "%s\n", "canceled.")
		return "", canceled, nil
	}}
}

// printChoice dumps the buffer and stops without acting on it, for when the
// user would rather take it from here by hand.
func printChoice(out io.Writer) choice {
	return choice{key: "p", label: "[p]rint", act: func(current string) (string, verdict, error) {
		printf(out, "%s\n", current)
		return "", canceled, nil
	}}
}
