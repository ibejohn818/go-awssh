package main

import (
	"fmt"
	"os"

	"github.com/ibejohn818/awssh/commands"
	"github.com/ibejohn818/awssh/config"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var vpConf = viper.New()
var cfgFile string

func MakeAwsshCli(conf *config.AwsshConf) *cobra.Command {

	cobra.OnInitialize(initConfig)

	conf.VpConf = vpConf

	cmd := &cobra.Command{
		Use:   "awssh",
		Short: "AWS SSH Tool",
		Long: `AWS SSH tool for ec2 instances plus more
	`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		//	Run: func(cmd *cobra.Command, args []string) { },
	}

	// flags := cmd.Flags()
	pFlags := cmd.PersistentFlags()

	pFlags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.awssh.yaml)")
	pFlags.StringVarP(&conf.Region, "region", "r", "us-west-2", "AWS Region")

	vpConf.BindPFlag("region", pFlags.Lookup("region"))

	commands.AddCommands(cmd, conf)

	return cmd
}

func main() {

	conf := &config.AwsshConf{}
	cmd := MakeAwsshCli(conf)

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		vpConf.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".awssh" (without extension).
		vpConf.AddConfigPath(home)
		vpConf.SetConfigName(".awssh")
		vpConf.SetConfigType("yaml")
	}

	// vpConf.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := vpConf.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", vpConf.ConfigFileUsed())
	}
}
