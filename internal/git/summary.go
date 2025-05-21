package git

import (
	"fmt"
	"sort"
	"strings"
)

type summary struct {
	diff  string
	stat  string
	names string
}

type change struct {
	Name        string
	Status      string // A, M, D, etc.
	StatLine    string
	DiffBlock   string
	LineChanges int
}

type report struct {
	Files      []change
	TotalLines int
}

func NewSummary(diff string, stat string, names string) *summary {
	return &summary{diff: diff, stat: stat, names: names}
}

func (s *summary) Render() string {
	rep := combine(s.names, s.stat, s.diff)
	return rep.pretty()
}

func (report *report) pretty() string {
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Name < report.Files[j].Name
	})
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total lines changed: %d\n", report.TotalLines))
	if report.large() {
		for _, file := range report.Files {
			sb.WriteString("\n\n")
			sb.WriteString(fmt.Sprintf("File: %s", file.Name))
			status := file.Status
			switch status {
			case "A":
				sb.WriteString(" was added")
				sb.WriteString(fmt.Sprintf("\nLines changed: %d\n", file.LineChanges))
				sb.WriteString("Diff block:\n")
				sb.WriteString(file.DiffBlock)
			case "D":
				sb.WriteString(" was removed")
			case "M":
				sb.WriteString(" was modified")
				sb.WriteString(fmt.Sprintf("\nLines changed: %d\n", file.LineChanges))
				sb.WriteString("Diff block:\n")
				sb.WriteString(file.DiffBlock)
			default:
				sb.WriteString(fmt.Sprintf(" has '%s' status", status))
			}
		}
	} else {
		for _, file := range report.Files {
			sb.WriteString("\n\n")
			sb.WriteString(fmt.Sprintf("File: %s", file.Name))
			status := file.Status
			switch status {
			case "A":
				sb.WriteString(" was added")
			case "D":
				sb.WriteString(" was removed")
			case "M":
				sb.WriteString(" was modified")
			default:
				sb.WriteString(fmt.Sprintf(" has '%s' status", status))
			}
			sb.WriteString(fmt.Sprintf("\nLines changed: %d\n", file.LineChanges))
			sb.WriteString("Diff block:\n")
			sb.WriteString(file.DiffBlock)
		}
	}
	return firstNLines(firstNChars(sb.String(), 120*400), 400)
}

func (report *report) large() bool {
	return report.TotalLines > 400
}

// Return only the first N lines of the report
func firstNLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// Return only the first N charactes of the report
func firstNChars(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

// Combine everything
func combine(names, stat, diff string) *report {
	statusMap := parseNames(names)
	statMap := parseStat(stat)
	diffMap := parseDiff(diff)
	report := &report{}
	for file, files := range statusMap {
		entry := change{
			Name:   file,
			Status: files,
		}
		if s, ok := statMap[file]; ok {
			entry.StatLine = s.StatLine
			entry.LineChanges = s.LineChanges
		}
		if d, ok := diffMap[file]; ok {
			entry.DiffBlock = d
		}
		report.Files = append(report.Files, entry)
		report.TotalLines += entry.LineChanges
	}
	return report
}

// Parse --name-status into map[file]status
// Input example:
//
// A	diff.txt
// A	simple.txt
// A	u1.txt
func parseNames(input string) map[string]string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	result := make(map[string]string)
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result[parts[1]] = parts[0]
		}
	}
	return result
}

// Parse --stat into map[file]FileChange (partial)
// Input example:
//
// diff.txt   | 1 +
// simple.txt | 7 +++++++
// u1.txt     | 7 +++++++
// 3 files changed, 15 insertions(+)
func parseStat(input string) map[string]*change {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	result := make(map[string]*change)
	for _, line := range lines {
		if !strings.Contains(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		file := strings.TrimSpace(parts[0])
		meta := strings.TrimSpace(parts[1])
		count := strings.Count(meta, "+") + strings.Count(meta, "-")
		result[file] = &change{
			Name:        file,
			StatLine:    line,
			LineChanges: count,
		}
	}
	return result
}

// Parse unified diff into map[file]diff
// Input example:
//
// diff --git a/diff.txt b/diff.txt
// new file mode 100644
// index 0000000..0bf3983
// --- /dev/null
// +++ b/diff.txt
// @@ -0,0 +1 @@
// +Hello diff
// diff --git a/simple.txt b/simple.txt
// new file mode 100644
// index 0000000..c4d7267
// --- /dev/null
// +++ b/simple.txt
// @@ -0,0 +1,7 @@
// +diff --git a/diff.txt b/diff.txt
// +new file mode 100644
// +index 0000000..0bf3983
// +--- /dev/null
// ++++ b/diff.txt
// +@@ -0,0 +1 @@
// ++Hello diff
// diff --git a/u1.txt b/u1.txt
// new file mode 100644
// index 0000000..c4d7267
// --- /dev/null
// +++ b/u1.txt
// @@ -0,0 +1,7 @@
// +diff --git a/diff.txt b/diff.txt
// +new file mode 100644
// +index 0000000..0bf3983
// +--- /dev/null
// ++++ b/diff.txt
// +@@ -0,0 +1 @@
// ++Hello diff
func parseDiff(input string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(input, "\n")
	var currentFile string
	var currentBlock []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "diff --git") {
			if currentFile != "" && len(currentBlock) > 0 {
				result[currentFile] = strings.Join(currentBlock, "\n")
			}
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				file := strings.TrimPrefix(parts[3], "b/")
				currentFile = file
				currentBlock = []string{line}
			}
		} else if currentFile != "" {
			currentBlock = append(currentBlock, line)
		}
	}
	if currentFile != "" && len(currentBlock) > 0 {
		result[currentFile] = strings.Join(currentBlock, "\n")
	}
	return result
}
