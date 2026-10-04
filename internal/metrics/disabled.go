package metrics

type disabled struct {
	origin Metrics
}

func NewDisabled(origin Metrics) Metrics {
	return &disabled{origin: origin}
}

func (d *disabled) Record(run Run) error {
	return nil
}

func (d *disabled) Runs() ([]Run, error) {
	return d.origin.Runs()
}
