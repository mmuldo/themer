package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/mmuldo/themer/theme"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new theme",
	Long: `Creates a new theme. Currently requires
an image to be specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		cvs, e := theme.GetColors(imgFile, 16)
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}

		p, e := theme.Delegate(cvs)
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}

		t, e := theme.Create(p, nil)
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}

		e = save(t, name)
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func save(theme *theme.Theme, name string) error {
	p := path.Join(os.Getenv("HOME"), ".config", "themer", "themes")

	if i, e := os.Stat(p); os.IsNotExist(e) || !i.IsDir() {
		os.MkdirAll(p, os.ModePerm)
	}

	json, e := json.MarshalIndent(*theme, "", "\t")
	if e != nil {
		return e
	}

	e = ioutil.WriteFile(path.Join(p, name), json, 0644)
	if e != nil {
		return e
	}

	return nil
}
