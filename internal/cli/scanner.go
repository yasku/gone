package cli

import (
	"bufio"
	"io"
)

type streamScanner struct {
	r   io.Reader
	s   *bufio.Scanner
	err error
}

func (ss *streamScanner) Scan() bool {
	if ss.err != nil {
		return false
	}
	if ss.s == nil {
		ss.s = bufio.NewScanner(ss.r)
		buf := make([]byte, 1024*1024)
		ss.s.Buffer(buf, 10*1024*1024)
	}
	return ss.s.Scan()
}

func (ss *streamScanner) Bytes() []byte {
	if ss.s == nil {
		return nil
	}
	return ss.s.Bytes()
}

func (ss *streamScanner) Err() error {
	if ss.err != nil {
		return ss.err
	}
	if ss.s == nil {
		return nil
	}
	return ss.s.Err()
}
