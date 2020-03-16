package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// availableCmd represents the available command
var availableCmd = &cobra.Command{
	Use:   "available",
	Short: "a list of available buffalo plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		plugs := Commands{
			{Name: authCmd.Use, BuffaloCommand: "generate", Description: authCmd.Short, Aliases: []string{}},
		}
		return json.NewEncoder(os.Stdout).Encode(plugs)
	},
}

func init() {
	RootCmd.AddCommand(availableCmd)
}

// Command that the plugin supplies
type Command struct {
	// Name "foo"
	Name string `json:"name"`
	// UseCommand "bar"
	UseCommand string `json:"use_command"`
	// BuffaloCommand "generate"
	BuffaloCommand string `json:"buffalo_command"`
	// Description "generates a foo"
	Description string   `json:"description,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Binary      string   `json:"-"`
	Flags       []string `json:"flags,omitempty"`
	// Filters events to listen to ("" or "*") is all events
	ListenFor string `json:"listen_for,omitempty"`
}

// Commands is a slice of Command
type Commands []Command
