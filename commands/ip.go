package commands

import (
	"fmt"

	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/config"
	"github.com/spf13/cobra"
)

func AddIpCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	private := false
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Get ip's in line by line list",
		Run: func(cmd *cobra.Command, args []string) {

			ec2Coll := api.GetServers(conf)

			ec2Res := ec2Coll.Filtered(args)

			for _, v := range ec2Res {

				ip := v.Ip

				if private {
					ip = v.PrivateIp
				}

				fmt.Println(ip)

			}

		},
	}

	flags := ipCmd.Flags()

	flags.BoolVarP(&private, "private", "p", false, "Private IP")

	cmd.AddCommand(ipCmd)
}
