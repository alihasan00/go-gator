package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBUrl               string `json:"db_url"`
	CurrentUserLoggedIn string `json:"current_user_name"`
}

func (c *Config) Read() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("something went wrong: %v", err)
		return err
	}

	filePath := homeDir + "/.gatorconfig.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("something went wrong: %v", err)
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		fmt.Printf("failed to parse config: %v", err)
		return err
	}
	return nil
}

func (c *Config) SetUser(current_user_name string) error {
	c.CurrentUserLoggedIn = current_user_name
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("something went wrong: %v", err)
		return err
	}

	filePath := homeDir + "/.gatorconfig.json"
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
