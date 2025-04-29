package pythonrt

import "bytes"

// Printer implements io.Writer and calls printFn for each line.
type Printer struct {
	buf     []byte
	printFn func(string) error
}

func (p *Printer) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	offset := 0
	for {
		i := bytes.IndexByte(p.buf[offset:], '\n')
		if i == -1 {
			break
		}

		line := string(p.buf[offset : offset+i])
		if err := p.printFn(line); err != nil {
			return 0, err
		}
		offset += i + 1
	}

	copy(p.buf, p.buf[offset:])
	p.buf = p.buf[:len(p.buf)-offset]

	return len(b), nil
}

func (p *Printer) Flush() error {
	if len(p.buf) > 0 {
		if err := p.printFn(string(p.buf)); err != nil {
			return err
		}
		p.buf = p.buf[:0]
	}
	return nil
}
