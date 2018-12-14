package commands

import (
	"fmt"

	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/config"
	"github.com/ibejohn818/awssh/shell"
	"github.com/spf13/cobra"
)

func AddSshCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	sshOpts := &shell.SSHOpts{}
	useBastion := false

	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH into an instance",
		Run: func(cmd *cobra.Command, args []string) {

			var instance api.Ec2Instance
			var bastion api.Ec2Instance

			ec2Coll := api.GetServers(conf)

			ec2Res := ec2Coll.Filtered(args)

			agent, err := shell.AgentAuth()
			if err != nil {
				fmt.Println("SSH Agent not detected/started in env SSH_AUTH_SOCK")
				sshOpts.IdentityFile = sshOpts.DefaultIdentityFile()
			} else {
				sshOpts.AuthMethods = append(sshOpts.AuthMethods, agent)
			}

			if len(sshOpts.IdentityFile) > 0 {

				sshKey, err := shell.NewSSHKey(sshOpts.IdentityFile)
				if err != nil {
					panic(err)
				}

				key, err := shell.SSHKeyAuth(sshKey)
				if err != nil {
					panic(err)
				}

				sshOpts.AuthMethods = append(sshOpts.AuthMethods, key)

			}

			if useBastion {
				bastion = selectInstance(ec2Res, "Select a bastion host")
			}

			instance = selectInstance(ec2Res, "Select an instance")

			sshClient := shell.NewSSHClient(instance, sshOpts)

			if len(bastion.Ip) > 0 {
				sshBastionClient := shell.NewSSHClient(bastion, sshOpts)
				shell.SSHBastionLogin(sshBastionClient, sshClient)
			} else {
				shell.SSHLogin(sshClient)
			}

		},
	}

	flags := sshCmd.Flags()
	flags.StringVarP(&sshOpts.User, "username", "u", "ec2-user", "SSH Username")
	conf.VpConf.BindPFlag("username", flags.Lookup("username"))

	// flags.StringVar(&sshSess.Command, "cmd", "", "Send command to the server")
	flags.BoolVarP(&sshOpts.Tty, "tty", "t", false, "TTY Terminal")
	flags.BoolVarP(&sshOpts.ForwardAuthSock, "auth", "A", false, "Pass SSH AUTH SOCK")
	flags.BoolVarP(&useBastion, "bastion", "b", false, "Connect via bastion host")
	flags.StringVarP(&sshOpts.IdentityFile, "identity", "i", "", "Path to ssh key file")
	flags.StringVarP(&sshOpts.Command, "cmd", "c", "", "Send command")
	flags.StringVarP(&sshOpts.Port, "port", "p", "22", "SSH Port")

	cmd.AddCommand(sshCmd)

}
