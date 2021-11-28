package handlers

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func (res *Result) SetCode(code int) *Result {
	res.Code = code
	return res
}

func (res *Result) UpdateAllFields(code int, msg string, data interface{}) *Result {
	res.SetCode(code)
	res.SetMessage(msg)
	res.SetData(data)
	return res
}

func (res *Result) SetMessage(msg string) *Result {
	res.Message = msg
	return res
}

func (res *Result) SetData(data interface{}) *Result {
	res.Data = data
	return res
}
