package commands

import (
	"fmt"

	"github.com/ibejohn818/go-awssh/api"
	"github.com/ibejohn818/go-awssh/config"
	"github.com/spf13/cobra"
)

func AddLsCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List servers",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {

			ec2Coll := api.GetServers(conf)

			ec2Res := ec2Coll.Filtered(args)

			for _, v := range ec2Res {
				fmt.Println(v.GetLine())
			}

		},
	}

	cmd.AddCommand(lsCmd)

}
