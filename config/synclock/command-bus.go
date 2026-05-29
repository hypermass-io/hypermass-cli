package synclock

import "sync"

// CommandRequest represents the data coming from the CLI
type CommandRequest struct {
	Command string
	Params  map[string]string
}

// CommandResponse is what we send back to the CLI
type CommandResponse struct {
	Success bool
	Message string
	Data    interface{}
}

type CommandBus struct {
	handlers map[string]func(CommandRequest) CommandResponse
	mu       sync.RWMutex
}

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]func(CommandRequest) CommandResponse),
	}
}

// Register allows a worker to claim a command name
func (b *CommandBus) Register(commandName string, handler func(CommandRequest) CommandResponse) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[commandName] = handler
}

// Dispatch is called by the ControlServer when an HTTP request arrives
func (b *CommandBus) Dispatch(req CommandRequest) CommandResponse {
	b.mu.RLock()
	handler, ok := b.handlers[req.Command]
	b.mu.RUnlock()

	if !ok {
		return CommandResponse{Success: false, Message: "Unknown command: " + req.Command}
	}
	return handler(req)
}
