package todo

const eventCreate = "create"
const eventUpdate = "update"
const eventDelete = "delete"

type Event struct {
	Event string `json:"event"`
	Todo  Todo   `json:"todo"`
}

func newEvent(event string, todo Todo) Event {
	return Event{Event: event, Todo: todo}
}

func newEventCreate(todo Todo) Event {
	return newEvent(eventCreate, todo)
}

func newEventUpdate(todo Todo) Event {
	return newEvent(eventUpdate, todo)
}

func newEventDelete(todo Todo) Event {
	return newEvent(eventDelete, todo)
}
