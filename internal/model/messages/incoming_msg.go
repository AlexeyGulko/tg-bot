package messages

type Commands map[string]Command

type Model struct {
	commands       Commands
	interceptor    Command
	defaultCommand Command
}

type Command interface {
	Execute(Message) bool
}

func New() *Model {
	return &Model{commands: make(Commands)}
}

func (s *Model) AddCommand(key string, command Command) {
	s.commands[key] = command
}

func (s *Model) SetDefaultCommand(command Command) {
	s.defaultCommand = command
}

func (s *Model) handleToInterceptor(msg Message) error {
	intercept := s.interceptor.Execute(msg)
	if !intercept {
		s.interceptor = nil
	}
	return nil
}

type Message struct {
	Text   string
	UserID int64
}

func (s *Model) IncomingMessage(msg Message) error {
	if s.interceptor != nil {
		return s.handleToInterceptor(msg)
	}

	comm, ok := s.commands[msg.Text]

	if !ok {
		s.defaultCommand.Execute(msg)
		return nil
	}

	intercept := comm.Execute(msg)
	if intercept {
		s.interceptor = comm
	}
	return nil
}
