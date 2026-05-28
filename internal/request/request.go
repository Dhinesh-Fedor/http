package request

import (
	"bytes"
	"fmt"
	"io"
	"my_http/internal/headers"
	"strconv"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers	 headers.Headers
	State       parserState
	Body          string
}

var (
	ERROR_BAD_START_LINE           = fmt.Errorf("Malformed Request-Line")
	ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported HTTP Version")
	ERROR_STATE                    = fmt.Errorf("Error in Request-Line")
)

var SEP = []byte("\r\n")

type parserState string

const (
	StateInit   parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
	StateHeaders parserState = "headers" 
	StateBody parserState = "body" 
)


func newRequest() *Request {
	return &Request{
		State: StateInit,
		Headers: *headers.NewHeaders(),
		Body: "",
	}
}

func (r *Request) hasBody() bool {
		clength := getInt(&r.Headers, "content-length",0)
		return  clength > 0
}

func (r *Request) done() bool {
	return r.State == StateDone || r.State == StateError
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)

	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value

}


func parseRequestLine(s []byte) (*RequestLine, int, error) {
	idx := bytes.Index(s, SEP)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := s[:idx]
	read := idx + len(SEP)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, nil
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_BAD_START_LINE
	}
  
	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {

		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		readN, err := request.parse(buf[:bufLen+n])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen+n])
		bufLen = bufLen+n - readN

	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.State {
		case StateError:
			return 0, ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.State = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.State = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)

			if err != nil{
				r.State = StateError
				return 0, err
			}

			if n == 0{
				break outer
			}

			read += n

			if done{
				if r.hasBody() {
					r.State = StateBody
				} else {
					r.State = StateDone
				}
			}

		case StateBody:
			clength := getInt(&r.Headers, "content-length",0)
			if clength == 0{
				panic("chuncked not implemented")
			}

			remain := min(clength - len(r.Body), len(currentData))
			r.Body += string(currentData[:remain])
			read += remain

			if len(r.Body) == clength{
				r.State = StateDone
			}
 
		case StateDone:
			break outer

		default:
			panic("somehow fked up!!")

		}
	}

	return read, nil
}

