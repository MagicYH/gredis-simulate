package proto

import (
	"gredissimulate/logger"
	"strconv"
	"strings"
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

// NewParser : Create a new parser
func NewParser() *Parser {
	return &Parser{
		cmd:           "",
		pState:        PARSE_CMD_COUNT,
		pCountCurrent: 0,
	}
}

// Request : request from client
type Request struct {
	Cmd    string
	Params []string
}

const RES_TYPE_STATE = "+"
const RES_TYPE_ERROR = "-"
const RES_TYPE_INT = ":"
const RES_TYPE_BULK = "$"
const RES_TYPE_MULTI = "*"

// SocketReader : Read line from socket
type SocketReader interface {
	ReadLine() (string, error)
}

// ResData : result data struct
type ResData struct {
	Type    string
	StrData string
}

// Response : response to client
type Response struct {
	Type string
	Data string
	Nest []*Response
}

// NewResponse : Create a new response
func NewResponse(t string) *Response {
	return &Response{Type: t, Data: ""}
}

// NewErrorRes : Create a new error response
func NewErrorRes(err string) *Response {
	return &Response{
		Type: RES_TYPE_ERROR,
		Data: err,
	}
}

// SetString : Set string
func (res *Response) SetString(content string) {
	res.Data = content
}

// SetInt : Set int
func (res *Response) SetInt(value int) {
	res.Data = strconv.Itoa(value)
}

// SetResponse :
func (res *Response) SetResponse(content *Response) {
	res.Nest = append(res.Nest, content)
}

// PARSE_CMD_COUNT : init, parse cmd count
const PARSE_CMD_COUNT = "parse_cmd_count"
const PARSE_PARAM_LEN = "parse_param_len"
const PARSE_PARAM = "parse_param"

const MSG_END = "\r\n"

// ParseError : ParseError
type ParseError struct {
	s string
}

// NewParseError : Create new ParseError
func NewParseError(s string) ParseError {
	return ParseError{
		s: s,
	}
}

func (e ParseError) Error() string {
	return e.s
}

// NetError : NetError
type NetError struct {
	s string
}

// NewNetError : Create new net error
func NewNetError(s string) NetError {
	return NetError{s: s}
}

func (e NetError) Error() string {
	return e.s
}

// ParseCmd : Parse command
func (parser *Parser) ParseCmd(reader SocketReader) (*Request, error) {
	for {
		content, err := reader.ReadLine()
		if nil != err {
			return nil, NewNetError(err.Error())
		}

		isOk, err := parser.DoParse(content)
		if nil != err {
			logger.LogError("Parse cmd error:", content)
			return nil, err
		}
		if isOk {
			break
		}
	}

	return parser.GetRequest(), nil
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
			e := parser.parseStrCmd(content)
			if nil == e {
				return true, nil
			}
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
			parser.cmd = strings.ToUpper(content)
		}
		if parser.pCountCurrent == parser.pCountTotal {
			return true, nil
		}
		parser.pState = PARSE_PARAM_LEN
	}
	return false, nil
}

// GetRequest : Get parse result of request
func (parser *Parser) GetRequest() *Request {
	return &Request{
		Cmd:    parser.cmd,
		Params: parser.params,
	}
}

// // BuildMultiResBinary : BuildMultiResBinary build multi response
// func BuildMultiResBinary(responseGroup *ResponseGroup) []byte {
// 	var content string
// 	if RESPONSE_GROUP_MULTI == responseGroup.Type {
// 		length := len(responseGroup.Responses)
// 		content = "*" + strconv.Itoa(length) + MSG_END
// 		for _, response := range responseGroup.Responses {
// 			content = content + buildResBinary(response)
// 		}
// 	} else {
// 		content = buildResBinary(responseGroup.Responses[0])
// 	}
// 	return []byte(content)
// }

// BuildResBinary : Convert response to binary result
func BuildResBinary(response *Response) string {
	var content string
	if RES_TYPE_MULTI == response.Type {
		content = "*" + strconv.Itoa(len(response.Nest)) + MSG_END
		for _, res := range response.Nest {
			content = content + BuildResBinary(res)
		}
	} else {
		if RES_TYPE_BULK == response.Type {
			strLen := len(response.Data)
			if 0 != strLen {
				content = response.Type + strconv.Itoa(strLen) + MSG_END + response.Data + MSG_END
			} else {
				content = response.Type + "-1" + MSG_END
			}
		} else {
			content = response.Type + response.Data + MSG_END
		}
	}

	return content
}

func parseCmdCount(content string) (int, error) {
	if (len(content) <= 0) || ("*" != content[0:1]) {
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

func (parser *Parser) parseStrCmd(content string) error {
	ss := strings.Split(content, " ")
	cmds := []string{}
	for _, s := range ss {
		if (" " != s) && ("" != s) {
			cmds = append(cmds, s)
		}
	}
	if len(cmds) <= 0 {
		return NewParseError("Parse cmd with text model fail")
	}

	parser.cmd = strings.ToUpper(cmds[0])
	for i := 1; i < len(cmds); i++ {
		parser.params = append(parser.params, cmds[i])
	}
	return nil
}
