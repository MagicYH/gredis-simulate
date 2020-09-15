package proto

import (
	"strconv"
)

// Parser : Protoc parser
type Parser struct {
	cmd           string
	params        []string
	pCountTotal   int
	pCountCurrent int
	pState        string
	pParamLen     int
}

// Request : request from client
type Request struct {
	Cmd    string
	Params []string
}

// Response : response to client
type Response struct {
	Type   string
	Params []string
}

// PARSE_CMD_COUNT : init, parse cmd count
const PARSE_CMD_COUNT = "parse_cmd_count"
const PARSE_PARAM_LEN = "parse_param_len"
const PARSE_PARAM = "parse_param"

// NewParser : Create a new parser
func NewParser() *Parser {
	return &Parser{
		cmd:           "",
		pState:        PARSE_CMD_COUNT,
		pCountCurrent: 0,
	}
}

// ParseError : ParseError
type ParseError struct {
	s string
}

// NewParseError : Create new ParseError
func NewParseError(s string) *ParseError {
	return &ParseError{
		s: s,
	}
}

func (e *ParseError) Error() string {
	return e.s
}

// DoParse : Parse command and parameters
//
// @param content string : content of input
//
// @return bool : if request is parse complete
// @return error : error occurred during parsing
func (parser *Parser) DoParse(content string) (bool, error) {
	switch parser.pState {
	case PARSE_CMD_COUNT:
		count, err := parseCmdCount(content)
		if nil != err {
			return false, NewParseError(err.Error())
		}
		parser.pCountTotal = count
		parser.pState = PARSE_PARAM_LEN

	case PARSE_PARAM_LEN:
		pLen, err := parseParamLen(content)
		if nil != err {
			return false, NewParseError(err.Error())
		}
		parser.pParamLen = pLen
		parser.pState = PARSE_PARAM

	case PARSE_PARAM:
		parser.pCountCurrent++
		if "" != parser.cmd {
			// If command is empty, then first parse command
			parser.params = append(parser.params, content)
		} else {
			parser.cmd = content
		}
		if parser.pCountCurrent == parser.pCountTotal {
			return true, nil
		}
	}
	return false, nil
}

// BuildResBinary : Convert response to binary result
func BuildResBinary(response *Response) []byte {
	return []byte("+PONG\r\n")
}

// GetRequest : Get parse result of request
func (parser *Parser) GetRequest() *Request {
	return &Request{
		Cmd:    parser.cmd,
		Params: parser.params,
	}
}

func parseCmdCount(content string) (int, error) {
	if "*" != content[0:1] {
		return 0, NewParseError("Cmd count proto error")
	}
	return strconv.Atoi(content[1:])
}

func parseParamLen(content string) (int, error) {
	if "$" != content[0:1] {
		return 0, NewParseError("Cmd param length proto error")
	}
	return strconv.Atoi(content[1:])
}
