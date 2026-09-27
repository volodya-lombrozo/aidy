package output

type Text func() (string, error)

func Fixed(text string) Text {
	return func() (string, error) {
		return text, nil
	}
}

type Output interface {
	Print(text Text) error
}
