// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/FactomProject/factoid/state"
	"github.com/FactomProject/factoid/state/stateinit"
	"strings"
)

type IState interface {
	GetCommand([]string) 	(ICommand, error)
	AddCommand(ICommand)
	Execute(args []string) 	error
	GetServer() 			string
	GetFS() 				state.IFactoidState
	GetCommands()			map[string]ICommand
}

type State struct {
	IState
	commands     map[string]ICommand
	fs           state.IFactoidState
	ipaddressFD  string
	portNumberFD string
	dbfile       string
}

var _ IState = (*State)(nil)

func (s State) GetCommands() map[string]ICommand {
	return s.commands
}

func (s State) GetFS() state.IFactoidState {
	return s.fs
}

func (s State) GetCommand(args []string) (ICommand, error) {
	if len(args) == 0 {
		return nil, nil
	}
	c := s.commands[strings.ToLower(args[0])]
	if c == nil {
		return nil, fmt.Errorf("Command not found")
	}
	return c, nil
}

func (s State) GetServer() string {
	return s.ipaddressFD + s.portNumberFD
}

func (s State) Execute(args []string) error {
	if len(args) == 0 {
		return nil
	}
	c, err := s.GetCommand(args)
	if err != nil {
		return err
	}
	return c.Execute(s, args)
}

func (s State) AddCommand(cmd ICommand) {
	s.commands[strings.ToLower(cmd.Name())] = cmd
}

func NewState(filename string) IState {
	s := new(State)

	s.dbfile = filename
	s.fs = stateinit.NewFactoidState(s.dbfile)
	s.commands = make(map[string]ICommand, 10)
	s.ipaddressFD = "localhost:"
	s.portNumberFD = "8088"

	s.AddCommand(new(AddFee))
	s.AddCommand(new(AddECOutput))
	s.AddCommand(new(AddressFromWords))
	s.AddCommand(new(AddInput))
	s.AddCommand(new(AddOutput))
	s.AddCommand(new(Balance))
	s.AddCommand(new(Balances))
	s.AddCommand(new(Export))
	s.AddCommand(new(ExportKey))
	s.AddCommand(new(Exit))
	s.AddCommand(new(Height))
	s.AddCommand(new(Help))
	s.AddCommand(new(Import))
	s.AddCommand(new(ImportKey))
	s.AddCommand(new(List))
	s.AddCommand(new(Listj))
	s.AddCommand(new(NewAddress))
	s.AddCommand(new(NewTransaction))
	s.AddCommand(new(Print))
	s.AddCommand(new(Run))
	//s.AddCommand(new(Setup))
	s.AddCommand(new(Sign))
	s.AddCommand(new(Submit))

	return s
}

/******************************************
 *  Command Interface
 ******************************************/

type ICommand interface {
	Execute(state IState, args []string) error
	Name() string
	ShortHelp() string // Short description
	LongHelp() string  // Detailed Help
}
