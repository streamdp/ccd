package domain

// Result structure for easily json serialization
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func NewResult(code int, msg string, data any) *Result {
	return &Result{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

// UpdateAllFields in Result and return it
func (r *Result) UpdateAllFields(code int, msg string, data any) *Result {
	r.Code = code
	r.Message = msg
	r.Data = data

	return r
}

// UpdateDataField in Result and return it
func (r *Result) UpdateDataField(data any) *Result {
	r.Data = data

	return r
}
