package main

import (
    "github.com/githubtofu/gator/internal/config"
    "fmt"
)

type command struct {
    name string
    args []string
}

type commands struct {
    handler map[string]func(*config.Config, command) error
}

func handlerLogin(s *config.Config, cmd command) error {
    if len(cmd.args) == 0 {
        fmt.Println("[command] No args provided")
        return fmt.Errorf("No arguments provided")
    }
    //s.CurrentUserName = cmd.args[0]
    s.SetUser(cmd.args[0])
    fmt.Println("User", cmd.args[0], "has logged in.")
    return nil
}

func (c *commands) register(name string, f func(*config.Config, command) error) {
    c.handler[name] = f
}

func (c *commands) runs(s *config.Config, cmd command) error {
    this_func, ok := c.handler[cmd.name]
    if !ok {
        fmt.Println("[command] no such command")
        return fmt.Errorf("No such command %v", cmd.name)
    }

    return this_func(s, cmd)
}
