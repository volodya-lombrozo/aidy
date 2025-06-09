package output

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter_Print(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printer := NewPrinter()

	_ = printer.Print("hello")
	err := w.Close()
	require.NoError(t, err, "failed to close write pipe")
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err, "failed to copy data")
	os.Stdout = originalStdout
	expected := "hello\n"
	require.Equal(t, expected, buf.String(), "Output should match expected value")
}
