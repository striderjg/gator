package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/striderjg/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdHandlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmdHandlers[name] = f
}
func (c *commands) run(s *state, cmd command) error {
	if _, ok := c.cmdHandlers[cmd.name]; !ok {
		return fmt.Errorf("command: %v does not exist", cmd.name)
	}
	return c.cmdHandlers[cmd.name](s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User: %v has been set\n", cmd.args[0])
	return nil
}

func main() {
	mainState := state{}
	var err error
	mainState.cfg, err = config.Read()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cmds := commands{
		cmdHandlers: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("Usage: gator requires a command in format gator COMMAND [ARGUMENTS]")
		os.Exit(1)
	}
	err = cmds.run(&mainState, command{os.Args[1], os.Args[2:]})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
