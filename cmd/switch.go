/*
Copyright © 2019 Matt Muldowney <matt.muldowney@gmail.com>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
)

var (
	terminal string
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		setDefaults()

		switch terminal {
		case "termite":
			template(path.Join(".config", "termite", "config"), name)
		default:
			fmt.Println(fmt.Errorf("'%s' is not a supported app", terminal))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)

	rootCmd.PersistentFlags().StringVarP(&terminal, "terminal", "t", "", "user terminal")
}

func setDefaults() {
	if terminal == "" {
		terminal = viper.GetString("terminal")
	}
}

func template(filepath string, theme string) error {
	ctxt := make(map[string]interface{})
	f, e := ioutil.ReadFile(path.Join(themesDir, theme))

	e = json.Unmarshal(f, &ctxt)

	tpl, e := pongo2.FromFile(path.Join("templates", filepath))
	if e != nil {
		return e
	}

	o, e := tpl.Execute(ctxt)
	if e != nil {
		return e
	}

	e = ioutil.WriteFile(path.Join(os.Getenv("HOME"), filepath), []byte(o), 0644)
	if e != nil {
		return e
	}

	return nil
}
