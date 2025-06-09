package output

import "fmt"

type printer struct {
}

func NewPrinter() Output {
	return &printer{}
}

func (p *printer) Print(command string) error {
	fmt.Println(command)
	return nil
}
