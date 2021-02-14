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

// Read populates buffer p with the next string from the stream.
func (ssr *stringStreamReader) Read(p []byte) (n int, err error) {

	nextString := <-ssr.stream

	// If there was an empty string put in the channel, we are done;
	// return an error
	if nextString == "" {
		return 0, io.EOF
	}

	return strings.NewReader(nextString).Read(p)
}

// Stream writes a string to the channel.
func (ssr *stringStreamReader) Stream(s string) {

	ssr.stream <- s
}

// Close will close the stream channel.
func (ssr *stringStreamReader) Close() {

	close(ssr.stream)
}
