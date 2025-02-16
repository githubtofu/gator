package config

import (
    "os"
    "encoding/json"
)

type Config struct {
    DbUrl string `json:"db_url"`
    CurrentUserName string `json:"current_user_name"`
}

func Read() Config {
    c := Config{}
    homedir, err := os.UserHomeDir()
    if err != nil {
        return c
    }
    f, err := os.Open(homedir + "/.gatorconfig.json")
    if err != nil {
        return c
    }
    if err := json.NewDecoder(f).Decode(&c); err != nil {
        return c
    }
    return c
}

func (c *Config) SetUser(u string) {
    c.CurrentUserName = u
    j_bytes, err := json.Marshal(c)
    if err != nil {
        return
    }
    homedir, err := os.UserHomeDir()
    if err := os.WriteFile(homedir + "/.gatorconfig.json", j_bytes, 0666); err != nil {
        return
    }
}

