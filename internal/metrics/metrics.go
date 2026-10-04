package metrics

import "time"

type Metrics interface {
	Record(run Run) error
	Runs() ([]Run, error)
}

type Run struct {
	Command  string    `json:"command"`
	Time     time.Time `json:"time"`
	Success  bool      `json:"success"`
	Duration int64     `json:"duration-ms"`
}

func NewRun(command string, start time.Time, err error) Run {
	return Run{
		Command:  command,
		Time:     start,
		Success:  err == nil,
		Duration: time.Since(start).Milliseconds(),
	}
}
