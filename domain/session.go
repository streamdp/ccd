package domain

type Session struct {
	Id       int64  `db:"_id"       json:"id"`
	TaskName string `db:"task_name" json:"task_name"`
	Interval int64  `db:"interval"  json:"interval"`
}
