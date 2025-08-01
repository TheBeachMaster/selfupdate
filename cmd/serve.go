/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	username  string
	repoName  string
	withToken bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the webserver",
	Long:  `Start a basic webserver`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
		flArgs := make(map[string]string, 0)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Shorthand != "h" { // Skip help
				flArgs["-"+f.Shorthand] = f.Value.String()
			}
		})

		flSb := strings.Builder{}

		for k, v := range flArgs {
			if v != "" {
				fmt.Fprintf(&flSb, "%s %s ", k, v)
			}
		}

		fmt.Printf("Flags %s\n", flSb.String())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&username, "username", "U", "TheBeachmaster", "Github Username for the owner of public repo you'd like to get stats on")
	serveCmd.Flags().StringVarP(&repoName, "repo", "R", "selfupdate", "Github public repo you'd like to get stats on")
	serveCmd.Flags().BoolVarP(&withToken, "withToken", "T", false, "Use Github token provided in .env")
}
