package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"

	"github.com/striderjg/gator/internal/config"
	"github.com/striderjg/gator/internal/database"
)

type state struct {
	db  *database.Queries
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

// ===================== HANDLERS ===============================================
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("user %v doesn't exist: %w", cmd.args[0], err)
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User: %v has been set\n", cmd.args[0])
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving users: %w", err)
	}
	for _, usr := range users {
		fmt.Printf("* %v", usr)
		if usr == s.cfg.Current_user {
			fmt.Println(" (current)")
		} else {
			fmt.Println("")
		}
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	if err := s.db.ClearDB(ctx); err != nil {
		return fmt.Errorf("error reseting the db: %w", err)
	}
	fmt.Println("Database reset")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the register command expects a single argument, the username")
	}
	ctx := context.Background()
	usr, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("error creating user %v: %w", cmd.args[0], err)
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User %v was created:\n", cmd.args[0])
	fmt.Printf("(id: %v, created_at: %v, updated_at: %v, name: %v\n", usr.ID, usr.CreatedAt, usr.UpdatedAt, usr.Name)
	return nil
}

// . ================================ ENTRY POINT ============================================
func main() {
	// --------- INIT
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

	db, err := sql.Open("postgres", mainState.cfg.Db_url)
	if err != nil {
		fmt.Printf("Error opening database %v\n", err.Error())
	}
	mainState.db = database.New(db)

	// -- Handler registration
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("users", handlerGetUsers)
	cmds.register("reset", handlerReset)

	// -- Start
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
