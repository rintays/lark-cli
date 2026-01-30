package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type Printer struct {
	Writer io.Writer
	JSON   bool
}

func (p Printer) Print(v any, text string) error {
	if p.JSON {
		enc := json.NewEncoder(p.Writer)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	_, err := fmt.Fprintln(p.Writer, text)
	return err
}
