package model

type Response struct {
	RequestId string      `json:"RequestId,omitempty"`
	Error     *ErrorInfo  `json:"Error,omitempty"`
	Data      interface{} `json:"Data,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"Code,omitempty"`
	Message string `json:"Message,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) SetData(data interface{}) {
	r.Data = data
}

func (r *Response) SetError(err ErrorInfo) {
	r.Error = &err
}

func (r *Response) SetRequestId(requestId string) {
	r.RequestId = requestId
}
