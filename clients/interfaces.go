package clients

// RestApiPuller interface makes it possible to expand the list of rest api pullers
type RestApiPuller interface {
	AddTask(from string, to string, interval int) *Task
	RemoveTask(from string, to string)
	Task(from string, to string) *Task
	ListTasks() Tasks
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
