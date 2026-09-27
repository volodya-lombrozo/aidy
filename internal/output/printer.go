package output

import "fmt"

type printer struct {
}

func NewPrinter() Output {
	return &printer{}
}

func (p *printer) Print(text Text) error {
	command, err := text()
	if err != nil {
		return err
	}
	fmt.Println(command)
	return nil
}
