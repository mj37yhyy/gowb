package model

type errorInfo struct {
	HttpStatus int
	Code       string `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
}

type Response struct {
	RequestId string      `json:"RequestId,omitempty"`
	Error     *errorInfo  `json:"Error,omitempty"`
	Data      interface{} `json:"Data,omitempty"`
}

type ErrorInfo struct {
	HttpStatus int
	Code       string
	Message    string
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) SetData(data interface{}) {
	r.Data = data
}

func (r *Response) SetError(err ErrorInfo) {
	r.Error = &errorInfo{err.Code, err.Message}
}

func (r *Response) SetRequestId(requestId string) {
	r.RequestId = requestId
}
