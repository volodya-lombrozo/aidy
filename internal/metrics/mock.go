package metrics

type mock struct {
	runs []Run
}

func NewMock(runs ...Run) Metrics {
	return &mock{runs: runs}
}

func (m *mock) Record(run Run) error {
	m.runs = append(m.runs, run)
	return nil
}

func (m *mock) Runs() ([]Run, error) {
	return m.runs, nil
}
