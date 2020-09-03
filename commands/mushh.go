package commands

import (
	"github.com/ibejohn818/go-awssh/config"
	"github.com/spf13/cobra"
)

func AddMusshCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	musshCmd := &cobra.Command{
		Use:   "mussh",
		Short: "Multi-SSH into an instance",
		Run: func(cmd *cobra.Command, args []string) {

			// spew.Dump(args)

			// ec2Coll := api.GetServers(conf)

			// ec2Res := ec2Coll.Filtered(args)

			// for k, v := range ec2Res {

			// 	ln := v.GetLine()
			// 	key := strconv.Itoa((k + 1))

			// 	fmt.Printf("%3v) %v \n", key, ln)

			// }

			// var instance api.Ec2Instance

			// prompt := promptui.Prompt{
			// 	Label: fmt.Sprintf("Choose a server [%v-%v]", strconv.Itoa(1), strconv.Itoa(len(ec2Res))),
			// }

			// ans, err := prompt.Run()

			// idx, err := strconv.Atoi(ans)

			// if err != nil || idx <= 0 || idx > len(ec2Res) {
			// 	fmt.Println("Invalid selection")
			// 	os.Exit(1)
			// }

			// instance = ec2Res[(idx - 1)]

			// musshOpts.Username = conf.VpConf.GetString("username")

			// sess := shell.NewSSH(instance, *musshOpts)

			// sess.Send()

		},
	}

	cmd.AddCommand(musshCmd)
}
