package csvee

import (
	"io"
	"strings"
)

type stringStreamReader struct {
	stream chan string
}

func newStringStreamReader() *stringStreamReader {

	return &stringStreamReader{
		stream: make(chan string),
	}
}

func (ssr *stringStreamReader) Read(p []byte) (n int, err error) {

	nextString := <-ssr.stream

	// If there was an empty string put in the channel, we are done;
	// return an error
	if nextString == "" {
		return 0, io.EOF
	}

	return strings.NewReader(nextString).Read(p)
}

func (ssr *stringStreamReader) Stream(s string) {

	ssr.stream <- s
}
