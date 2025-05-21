package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Diff = `diff --git a/diff.txt b/diff.txt
new file mode 100644
index 0000000..0bf3983
--- /dev/null
+++ b/diff.txt
@@ -0,0 +1 @@
+Hello diff
diff --git a/simple.txt b/simple.txt
new file mode 100644
index 0000000..c4d7267
--- /dev/null
+++ b/simple.txt
@@ -0,0 +1,7 @@
+diff --git a/diff.txt b/diff.txt
+new file mode 100644
+index 0000000..0bf3983
+--- /dev/null
++++ b/diff.txt
+@@ -0,0 +1 @@
++Hello diff
diff --git a/u1.txt b/u1.txt
new file mode 100644
index 0000000..c4d7267
--- /dev/null
+++ b/u1.txt
@@ -0,0 +1,7 @@
+diff --git a/diff.txt b/diff.txt
+new file mode 100644
+index 0000000..0bf3983
+--- /dev/null
++++ b/diff.txt
+@@ -0,0 +1 @@
++Hello diff`
	Pretty = `Total lines changed: 15


File: diff.txt was added
Lines changed: 1
Diff block:
diff --git a/diff.txt b/diff.txt
new file mode 100644
index 0000000..0bf3983
--- /dev/null
+++ b/diff.txt
@@ -0,0 +1 @@
+Hello diff

File: simple.txt was added
Lines changed: 7
Diff block:
diff --git a/simple.txt b/simple.txt
new file mode 100644
index 0000000..c4d7267
--- /dev/null
+++ b/simple.txt
@@ -0,0 +1,7 @@
+diff --git a/diff.txt b/diff.txt
+new file mode 100644
+index 0000000..0bf3983
+--- /dev/null
++++ b/diff.txt
+@@ -0,0 +1 @@
++Hello diff

File: u1.txt was added
Lines changed: 7
Diff block:
diff --git a/u1.txt b/u1.txt
new file mode 100644
index 0000000..c4d7267
--- /dev/null
+++ b/u1.txt
@@ -0,0 +1,7 @@
+diff --git a/diff.txt b/diff.txt
+new file mode 100644
+index 0000000..0bf3983
+--- /dev/null
++++ b/diff.txt
+@@ -0,0 +1 @@
++Hello diff`
)

func TestSummary_Render(t *testing.T) {
	summary := NewSummary(
		Diff,
		`diff.txt   | 1 +
        simple.txt | 7 +++++++
        u1.txt     | 7 +++++++
        3 files changed, 15 insertions(+)`,
		`A	diff.txt
        A	simple.txt
        A	u1.txt`,
	)
	res := summary.Render()

	assert.Equal(t, Pretty, res)
}

func TestTrimToNChars_MoreThan400Lines(t *testing.T) {
	input := generateLongReport(500, 10)
	trimLength := 400

	result := firstNLines(input, trimLength)

	assert.Equal(t, 400, lines(result))
}

func TestTrimToNChars_MoreThan120x400chars(t *testing.T) {
	input := generateLongReport(50, 100)
	trimLength := 10

	result := firstNChars(input, trimLength)

	assert.Equal(t, 10, len(result))
}

func generateLongReport(lines int, charsPerLine int) string {
	line := strings.Repeat("a", charsPerLine)
	return strings.Repeat(line+"\n", lines)
}

func lines(s string) int {
	return strings.Count(s, "\n") + 1
}
