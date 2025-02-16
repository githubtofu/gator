package main
import _ "github.com/lib/pq"

import (
    "github.com/githubtofu/gator/internal/config"
    "github.com/githubtofu/gator/internal/database"
    "github.com/google/uuid"
    "os"
    "fmt"
    "time"
    "context"
    "database/sql"
)

type state struct {
    c *config.Config
    db *database.Queries
}

type command struct {
    name string
    args []string
}

type commands struct {
    handler map[string]func(state, command) error
}

func handlerRegister(st state, cmd command) error {
    if len(cmd.args) == 0 {
        fmt.Println("[command] No args provided")
        return fmt.Errorf("No arguments provided")
    }
    p := database.CreateUserParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name: cmd.args[0],
    }
    
    u, err := st.db.CreateUser(context.Background(), p)
    if err != nil {
        fmt.Printf("%w", err)
        os.Exit(1)
    }
    st.c.SetUser(u.Name)
    fmt.Println("User", u.Name, "has been registered.")
    return nil
}

func handlerLogin(st state, cmd command) error {
    if len(cmd.args) == 0 {
        fmt.Println("[command] No args provided")
        return fmt.Errorf("No arguments provided")
    }
    u, err := st.db.GetUser(context.Background(), cmd.args[0])
    if err != nil {
        fmt.Printf("%w", err)
        os.Exit(1)
    }
    st.c.SetUser(u.Name)
    fmt.Println("User", u.Name, "has logged in.")
    return nil
}

func (c commands) register(name string, f func(state, command) error) {
    c.handler[name] = f
}

func (c commands) runs(st state, cmd command) error {
    this_func, ok := c.handler[cmd.name]
    if !ok {
        fmt.Println("[command] no such command")
        return fmt.Errorf("No such command %v", cmd.name)
    }

    return this_func(st, cmd)
}

func main() {
    st := state{}
    st.c = config.Read()
    dbURL := st.c.DbUrl
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    }
    st.db = database.New(db)
    m_commands := commands{}
    m_commands.handler = make(map[string]func(state, command) error)
    m_commands.register("login", handlerLogin)
    m_commands.register("register", handlerRegister)
    args :=os.Args
    if len(args) < 2 {
        fmt.Println("Command needed")
        os.Exit(1)
    }
    m_command := command{name:args[1], args:args[2:]}
    if err := m_commands.runs(st, m_command); err != nil {
        fmt.Printf("%w", err)
        os.Exit(1)
    } 
}
