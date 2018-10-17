package kuu

// EventHandler 事件处理器
type EventHandler func(...interface{})

// EventHandlers 事件处理器集合
type EventHandlers []EventHandler

var events = map[string]EventHandlers{}

// On 事件订阅
func On(event string, handlers ...EventHandler) {
	if event == "" || len(handlers) == 0 {
		return
	}
	for _, handler := range handlers {
		if events[event] == nil {
			events[event] = make(EventHandlers, 0)
		}
		events[event] = append(events[event], handler)
	}
}

// Emit 事件触发
func Emit(event string, args ...interface{}) {
	if event == "" || events[event] == nil {
		return
	}
	handlers := events[event]
	for _, handler := range handlers {
		handler(args...)
	}
}
