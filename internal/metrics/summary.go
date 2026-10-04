package metrics

import (
	"sort"
	"time"
)

type Summary struct {
	Command string
	Runs    int
	Failed  int
	Last    time.Time
}

func Summarize(runs []Run, commands []string) []Summary {
	summaries := make(map[string]*Summary, len(commands))
	for _, command := range commands {
		summaries[command] = &Summary{Command: command}
	}
	for _, run := range runs {
		summary, ok := summaries[run.Command]
		if !ok {
			summary = &Summary{Command: run.Command}
			summaries[run.Command] = summary
		}
		summary.add(run)
	}
	result := make([]Summary, 0, len(summaries))
	for _, summary := range summaries {
		result = append(result, *summary)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Runs != result[j].Runs {
			return result[i].Runs > result[j].Runs
		}
		return result[i].Command < result[j].Command
	})
	return result
}

func (s *Summary) add(run Run) {
	s.Runs++
	if !run.Success {
		s.Failed++
	}
	if run.Time.After(s.Last) {
		s.Last = run.Time
	}
}
