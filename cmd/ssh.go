package cmd

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/shell"
	"github.com/spf13/cobra"
)

type SshOps struct {
	User       string
	UseBastion bool
	PubKeyPath string
	UseEc2Conn bool
	Port       string
}

func AddSshCmd(aCmd *cobra.Command, gops *GlobalConfig) *cobra.Command {

	ops := SshOps{}

	sshCmd := cobra.Command{
		Use:   "ssh",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {

			var bastionClient *shell.SSHClient

			api.Region = gops.Region

			target, bastion, selErr := selectInstances(&ops)

			if selErr != nil {
				log.Fatal("Invalid choice")
			}
			sshConnOps := shell.NewSSHOpts()
			spew.Dump(sshConnOps)
			spew.Dump(target)
			spew.Dump(bastion)
			spew.Dump(bastionClient)
		},
	}

	flags := sshCmd.Flags()

	flags.StringVarP(&ops.User, "user", "u", "", "SSH Username to use")
	flags.StringVarP(&ops.Port, "port", "p", "22", "SSH port to connect with")
	flags.StringVarP(&ops.PubKeyPath, "identity", "i", "", "Path to ssh public key to use")
	flags.BoolVarP(&ops.UseBastion, "bastion", "b", false, "Connect via a bastion host")
	flags.BoolVarP(&ops.UseEc2Conn, "ec2connect", "c", false, "Send public key via ec2-instance-connect")
	aCmd.AddCommand(&sshCmd)

	return &sshCmd

}

func sendSSHKeys(target *api.Ec2Instance, bastion *api.Ec2Instance, ops *SshOps) {

	if !ops.UseEc2Conn {
		return
	}

}

func selectInstances(ops *SshOps) (*api.Ec2Instance, *api.Ec2Instance, error) {
	var bastionTarget *api.Ec2Instance
	var target *api.Ec2Instance
	var eErr error
	ec2Client := api.NewEc2Client()

	list := ec2Client.GetInstances()

	if ops.UseBastion {
		var bErr error

		bastionTarget, bErr = api.SelectInstance(list, "Select Bastion Host")

		if bErr != nil {
			log.Fatal("Invalid bastion instance")
		}
	}

	target, eErr = api.SelectInstance(list, "Select instance")

	if eErr != nil {

	}

	return target, bastionTarget, nil
}
