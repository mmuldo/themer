/*
Copyright © 2019 Matt Muldowney <matt.muldowney@gmail.com>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
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
		for _, v := range user_apps {
			e := template(supported_apps[v], name)
			if e != nil {
				fmt.Println(e)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)

	pongo2.RegisterFilter("hex2RGB", filterHex2RGB)
}

func hex2RGB(hex string) (int, int, int, error) {
	h := strings.Split(hex, "#")
	hex = h[len(h)-1]
	if len(hex) != 6 {
		return -1, -1, -1, fmt.Errorf("%s is not a valid hex value.", hex)
	}

	r, e := strconv.ParseUint(hex[:2], 16, 8)
	if e != nil {
		return -1, -1, -1, e
	}

	g, e := strconv.ParseUint(hex[2:4], 16, 8)
	if e != nil {
		return -1, -1, -1, e
	}

	b, e := strconv.ParseUint(hex[4:], 16, 8)
	if e != nil {
		return -1, -1, -1, e
	}

	return int(r), int(g), int(b), nil
}

var filterHex2RGB pongo2.FilterFunction = func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	r, g, b, e := hex2RGB(in.String())
	if e != nil {
		err := pongo2.Error{
			nil, "", -1, -1, nil, "", e,
		}
		fmt.Println(e)
		return nil, &err
	}

	return pongo2.AsValue([]int{r, g, b}), nil
}

func template(filepath string, theme string) error {
	fmt.Println(theme)
	ctxt := make(map[string]interface{})
	f, e := ioutil.ReadFile(path.Join(themesDir, theme))
	if f == nil {
		fmt.Println("failed")
		os.Exit(1)
	}
	if e != nil {
		return e
	}

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
