package airworthy

import (
	"fmt"
	"io"
)

type Progress struct {
	Reader io.ReadCloser
	Size   int64 // expected size

	readTotal int64 // bytes read
}

func (p *Progress) Read(d []byte) (int, error) {
	n, err := p.Reader.Read(d)
	p.readTotal += int64(n)
	return n, err
}

func (p *Progress) Progress() string {
	return fmt.Sprintf("%d out of %d read", p.readTotal, p.Size)
}
