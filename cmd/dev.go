package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/api"
	"github.com/spf13/cobra"
)

func do_stuff() int {
	return 1
}

// AddDevCmd ....
func AddDevCmd(aCmd *cobra.Command, gops *GlobalConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dev",
		Run: func(cmd *cobra.Command, args []string) {

			api.Region = gops.Region
			ec2 := api.NewEc2Client()
			ec2Conn := api.NewEc2ConnClient()

			list := ec2.GetInstances()
			// for _, i := range inst {
			// 	spew.Dump(i.GetTplMap())
			// }

			inst, _ := api.SelectInstance(list)

			spew.Dump(inst)

			payload := api.Ec2ConnPayload{
				User:       "danb",
				Instance:   *inst,
				PubKeyPath: "/Users/jhardy/.ssh/id_rsa.pub",
			}

			ec2Conn.SendPublicKey(&payload)

			// tp := utils.NewPrompt(func(p *utils.TextPrompt) {
			// 	p.InputBuffer = os.Stdin
			// })

			// ans, _ := tp.Ask("Whay is your name?")

			// spew.Dump(ans)

			// ec2 := api.NewEc2Client()

			// res := ec2.GetInstances()

			// spew.Dump(res)
		},
	}

	aCmd.AddCommand(cmd)
	return cmd
}
