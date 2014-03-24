package channeldispatch

import (
	"fmt"
	"sync"
)

type event struct {
	key  string
	data interface{}
	ret  chan interface{}
}

func (ev *event) Param() interface{} { return ev.data }

func (ev *event) Return(val interface{}) {
	ev.ret <- val
}

type Event interface {
	Param() interface{}
	Return(interface{})
	Returned() interface{}
}

func (ev *event) Returned() interface{} {
	data := <-ev.ret
	close(ev.ret)
	return data
}

var events = make(chan *event, 20)
var eventHandlers = struct {
	*sync.RWMutex
	handlers map[string]func(Event)
}{
	&sync.RWMutex{},
	map[string]func(Event){},
}

func New(s string) EventString {
	return EventString(s)
}

type EventString string

func (e EventString) Throw(param interface{}) Event {
	ret := make(chan interface{}, 1)
	ev := &event{
		key:  string(e),
		data: param,
		ret:  ret,
	}
	events <- ev
	return ev
}

func (e EventString) Handle(handler func(Event)) {
	eventHandlers.Lock()
	defer eventHandlers.Unlock()
	eventHandlers.handlers[string(e)] = handler
}

func (e EventString) UnHandle() {
	eventHandlers.Lock()
	defer eventHandlers.Unlock()
	delete(eventHandlers.handlers, string(e))
}

func callHandler(ev *event) {
	eventHandlers.RLock()
	handler := eventHandlers.handlers[ev.key]
	eventHandlers.RUnlock()
	if handler == nil {
		fmt.Println("unhandled event: " + ev.key)
		return
	}
	if handler != nil {
		handler(ev)
		// handler.ServeHTTP(ev.ResponseWriter, ev.Request)
	}
}

func Watch() {
	go func() {
		for {
			select {
			case ev := <-events:
				callHandler(ev)
				// default:
			}
		}
	}()
}
