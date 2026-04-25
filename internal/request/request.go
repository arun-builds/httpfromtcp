package request

import (
	"bytes"
	"strconv"

	"fmt"
	"io"

	"github.com/arun-builds/httpfromtcp/internal/headers"
)

type parserState string

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
	Body 	string
}

func getInt(headers *headers.Headers, name string, defaultValue int)int{
	valueStr, exists := headers.Get(name)
	if !exists{
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil{
		return defaultValue
	}

	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body: "",
	}
}

var ERROR_BAD_REQUEST_LINE = fmt.Errorf("bad request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_REQUEST_ERROR_STATE = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

const (
	StateInit    parserState = "initialized"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, nil
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_BAD_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool{
	// TODO: when doing chunked encoding, update this method
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0{
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n

			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			// in real world we would not get EOF after reading data
			// therefor we would nicely transition to body, whichc would allow
			// us to then transition to don, but doing the transition here
			if done {
				if r.hasBody(){
				r.state = StateBody
				}else{
					r.state = StateDone
				}
			}
		
		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0{
				panic("chunked not implemented")
			}

			remaining := min(length - len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length{
				r.state = StateDone
			}
		
		case StateDone:
			break outer

		default:
			panic("errooooor")
		}

	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := newRequest()

	// Note: buffer could overrun, a header that exceeds 1k or the body
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		// TODO: resolve err
		if err != nil {
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN

	}

	return request, nil
}
