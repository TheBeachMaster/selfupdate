/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"com.thebeachmaster/selfupdate/internal/config"
	"com.thebeachmaster/selfupdate/internal/pkg/version"
	"com.thebeachmaster/selfupdate/internal/service"
	backOffStrategy "github.com/cenkalti/backoff/v5"
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

		// _user := cmd.Flag("username").Value.String()
		// _repo := cmd.Flag("repo").Value.String()
		_withTkn := cmd.Flag("withToken").Value.String()
		m_withTkn := _withTkn == "true"

		fmt.Printf("Flags %s\n", flSb.String())
		fmt.Printf("Starting service v%s \n", version.CurrentVersion)

		// Muxer
		mux := http.NewServeMux()

		// Check if port is available then start

		if _, err := backOffStrategy.Retry(context.Background(), func() (bool, error) {
			_addr := ":8098"
			_listener, err := net.Listen("tcp", _addr)
			if err != nil {
				log.Printf("port is in use %s \n", err.Error())
				return false, fmt.Errorf("port in use")
			}

			_ = _listener.Close()
			opts := make([]string, 0)
			if m_withTkn {
				_cfg := config.New()
				p_cfg, err := _cfg.Parse()
				if err != nil {
					log.Printf("invalid config error %s\n", err.Error())
					return false, fmt.Errorf("invalid config")
				}
				opts = append(opts, p_cfg.Github.AcccessToken)
			}

			_hdlr := service.NewServiceHandler(flSb.String(), opts...)
			mux.Handle("/version", _hdlr.CheckAppVersionHandler())
			mux.Handle("/update", _hdlr.UpdateAppHandler())
			go func() {
				log.Printf("INFO: starting server on %s...\n", _addr)
				if err := http.ListenAndServe(_addr, mux); err != nil {
					log.Printf("ERROR: server error %s \n", err.Error())
					panic(err)
				}
			}()

			log.Println("INFO: server started")
			return true, nil
		}, backOffStrategy.WithBackOff(backOffStrategy.NewExponentialBackOff()), backOffStrategy.WithMaxElapsedTime(time.Second*15)); err != nil {
			log.Printf("ERROR: failed to start service on time due to %s", err)
			os.Exit(1)
		}

		quitServer := make(chan struct{})

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		close(quitServer)

		<-quitServer
		log.Println("server shutdown")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&username, "username", "U", "TheBeachmaster", "Github Username for the owner of public repo you'd like to get stats on")
	serveCmd.Flags().StringVarP(&repoName, "repo", "R", "selfupdate", "Github public repo you'd like to get stats on")
	serveCmd.Flags().BoolVarP(&withToken, "withToken", "T", false, "Use Github token provided in .env")
}
