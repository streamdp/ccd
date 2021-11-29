package handlers

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func (res *Result) UpdateAllFields(code int, msg string, data interface{}) *Result {
	res.Code = code
	res.Message = msg
	res.Data = data
	return res
}

func (res *Result) UpdateDataField(data interface{}) *Result {
	res.Data = data
	return res
}
