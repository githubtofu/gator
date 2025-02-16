package main

import (
    "github.com/githubtofu/gator/internal/config"
    "os"
    "fmt"
)

func main() {
    c := config.Read()
    m_commands := commands{}
    m_commands.handler = make(map[string]func(*config.Config, command) error)
    m_commands.register("login", handlerLogin)
    args :=os.Args
    if len(args) < 2 {
        fmt.Println("Command needed")
        os.Exit(1)
    }
    m_command := command{name:args[1], args:args[2:]}
    err := m_commands.runs(&c, m_command)
    if err != nil {
        fmt.Printf("%w", err)
        os.Exit(1)
    } 
}
