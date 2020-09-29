/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	homedir "github.com/mitchellh/go-homedir"
)

type GlobalConfig struct {
	Region string
}

// rootCmd represents the base command when called without any subcommands
var (
	cfgFile string
	rootCmd *cobra.Command
)

func InitCli() (*cobra.Command, *GlobalConfig) {
	globalOps := GlobalConfig{}

	awsshCmd := cobra.Command{
		Use:   "awssh",
		Short: "AWS Ec2 ssh connections",
		//Run: func(cmd *cobra.Command, args []string) {},

	}

	pflags := awsshCmd.PersistentFlags()

	awsshCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.awssh.yaml)")
	pflags.StringVarP(&globalOps.Region, "region", "r", "us-west-2", "AWS Region")

	return &awsshCmd, &globalOps
}

// AddCommands ....
func addCommands(cmd *cobra.Command, ops *GlobalConfig) {
	// AddDevCmd(cmd, ops)
	AddSshCmd(cmd, ops)
	AddLsCmd(cmd, ops)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	_rootCmd, gops := InitCli()

	rootCmd = _rootCmd

	cobra.OnInitialize(initConfig)

	// attach sub commands
	addCommands(rootCmd, gops)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".awssh" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".awssh")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
