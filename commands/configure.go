package commands

import (
	"fmt"

	"github.com/ibejohn818/go-awssh/config"
	"github.com/spf13/cobra"
)

func AddConfigureCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	configureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure and save defaults",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, World!")
		},
	}

	cmd.AddCommand(configureCmd)
}
