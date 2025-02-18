package main
import _ "github.com/lib/pq"

import (
    "log"
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
    handler map[string]func(*state, command) error
}

func handlerRegister(st *state, cmd command) error {
    if len(cmd.args) == 0 {
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
        fmt.Errorf("User already exists. %w", err)
        os.Exit(1)
    }
    st.c.SetUser(u.Name)
    fmt.Println("User", u.Name, "has been registered.")
    return nil
}

func handlerUsers(st *state, cmd command) error {
    us, err := st.db.GetUsers(context.Background())
    if err != nil {
        fmt.Errorf("Problem getting users. %w", err)
    }
    for _, v := range us {
        displayed_user := v.Name
        if v.Name == st.c.CurrentUserName {
            displayed_user += " (current)"
        }
        fmt.Printf("* %v\n", displayed_user)
    }
    return nil
}

func handlerReset(st *state, cmd command) error {
    err := st.db.Reset(context.Background())
    if err != nil {
        fmt.Errorf("Problem resetting. %w", err)
    }
    fmt.Println("Successfully reset")
    return nil
}

func handlerLogin(st *state, cmd command) error {
    if len(cmd.args) == 0 {
        return fmt.Errorf("No arguments provided")
    }
    u, err := st.db.GetUser(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("couldn't find user: %w", err)
    }
    st.c.SetUser(u.Name)
    fmt.Println("User", u.Name, "has logged in.")
    return nil
}

func (c commands) register(name string, f func(*state, command) error) {
    c.handler[name] = f
}

func (c commands) runs(st *state, cmd command) error {
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
    m_commands.handler = make(map[string]func(*state, command) error)
    m_commands.register("login", handlerLogin)
    m_commands.register("register", handlerRegister)
    m_commands.register("reset", handlerReset)
    m_commands.register("users", handlerUsers)
    m_commands.register("agg", handlerAgg)
    m_commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
    m_commands.register("feeds", handlerFeeds)
    m_commands.register("following", middlewareLoggedIn(handlerFollowing))
    m_commands.register("follow", middlewareLoggedIn(handlerFollow))
    m_commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
    args :=os.Args
    if len(args) < 2 {
        fmt.Println("Command needed")
        os.Exit(1)
    }
    m_command := command{name:args[1], args:args[2:]}
    fmt.Println("Running", m_command.name)
    if err := m_commands.runs(&st, m_command); err != nil {
        log.Fatal(err)
    } 
}
