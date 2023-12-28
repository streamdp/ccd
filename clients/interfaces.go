package clients

// RestApiPuller interface makes it possible to expand the list of rest api pullers
type RestApiPuller interface {
	Task(from string, to string) *Task
	AddTask(from string, to string, interval int64) *Task
	RemoveTask(from string, to string)
	ListTasks() Tasks
	UpdateTask(t *Task, interval int64) *Task
	RestoreLastSession() error
}

// RestClient interface makes it possible to expand the list of rest data providers
type RestClient interface {
	Get(from string, to string) (*Data, error)
}

// WsClient interface makes it possible to expand the list of wss data providers
type WsClient interface {
	Subscribe(from string, to string) error
	Unsubscribe(from string, to string) error
	ListSubscribes() Subscribes
}
