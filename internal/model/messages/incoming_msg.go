package messages

import (
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
)

type Commands map[string]Command

type Model struct {
	commands       Commands
	defaultCommand Command
	stopCommand    Command
	commandStorage Storage
}

type Command interface {
	Execute(context.Context, *dto.Message) CommandError
	Next() (Command, bool)
}

type Storage interface {
	Add(int64, Command)
	Get(int64) (Command, bool)
	Delete(int64)
}

type CommandError interface {
	error
	DoRetry() bool
}

func New(storage Storage) *Model {
	return &Model{commands: make(Commands), commandStorage: storage}
}

func (s *Model) AddCommand(key string, command Command) {
	s.commands[key] = command
}

func (s *Model) SetDefaultCommand(command Command) {
	s.defaultCommand = command
}

func (s *Model) SetStopCommand(command Command) {
	s.stopCommand = command
}

func (s *Model) IncomingMessage(ctx context.Context, msg *dto.Message) error {
	if msg.Text == "/stop" {
		err := s.stopCommand.Execute(ctx, msg)
		s.commandStorage.Delete(msg.UserID)
		if err != nil {
			return err
		}
		return nil
	}

	var command Command
	hasCommand := false
	stored := false

	command, stored = s.commandStorage.Get(msg.UserID)

	if !stored {
		command, hasCommand = s.commands[msg.Text]
	}

	if !hasCommand && !stored {
		err := s.defaultCommand.Execute(ctx, msg)
		return err
	}

	err := command.Execute(ctx, msg)
	if err != nil {
		if err.DoRetry() {
			return nil
		}
		s.commandStorage.Delete(msg.UserID)
		return err
	}

	if v, has := command.Next(); has {
		s.commandStorage.Add(msg.UserID, v)
	} else {
		if stored {
			s.commandStorage.Delete(msg.UserID)
		}
	}
	return nil
}
