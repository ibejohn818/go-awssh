package cmd

import (
	"fmt"
	"log"

	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/shell"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
)

type SshOps struct {
	User          string
	UseBastion    bool
	UsePrivateIp  bool
	PrivKey       string
	PubKey        string
	UseEc2Conn    bool
	Port          string
	ForwardSocket bool
}

func AddSshCmd(aCmd *cobra.Command, gops *GlobalConfig) *cobra.Command {

	ops := SshOps{}

	sshCmd := cobra.Command{
		Use:   "ssh",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {

			api.Region = gops.Region

			target, bastion, selErr := selectInstances(&ops)

			if selErr != nil {
				log.Fatal("Invalid choice")
			}

			sshConnOps := shell.NewSSHOpts(func(o *shell.SSHOpts) {
				if len(ops.User) > 0 {
					o.User = ops.User
				}
				if len(ops.PrivKey) > 0 {
					o.IdentityFile = ops.PrivKey
				}
				if len(ops.Port) > 0 {
					o.Port = ops.Port
				}
				if ops.ForwardSocket {
					o.ForwardAuthSock = true
				}

				if ops.UsePrivateIp {
					o.UsePrivateIp = true
				} else {
					o.UsePrivateIp = false
				}
			})

			handleSshConn(target, bastion, &ops, &sshConnOps)

		},
	}

	flags := sshCmd.Flags()

	flags.StringVarP(&ops.User, "user", "u", "", "SSH Username to use")
	flags.StringVarP(&ops.Port, "port", "p", "22", "SSH port to connect with")
	flags.StringVarP(&ops.PrivKey, "identity", "i", "", "Path to ssh private key to use")
	flags.BoolVarP(&ops.UseBastion, "bastion", "b", false, "Connect via a bastion host")
	flags.BoolVarP(&ops.UseEc2Conn, "ec2connect", "c", false, "Send public key via ec2-instance-connect")
	flags.BoolVarP(&ops.ForwardSocket, "auth", "A", false, "Forward SSH_AUTH_SOCK")
	flags.BoolVarP(&ops.UsePrivateIp, "private", "", false, "Use private ip address")
	aCmd.AddCommand(&sshCmd)

	return &sshCmd

}

func handleSshConn(target *api.Ec2Instance, bastion *api.Ec2Instance, ops *SshOps, sshOpts *shell.SSHOpts) {

	if ops.UseEc2Conn {
		sendSSHKeys(target, bastion, ops, sshOpts)
	}

	targetClient := shell.NewSSHClient(*target, sshOpts)

	if ops.UseBastion {
		bastionClient := shell.NewSSHClient(*bastion, sshOpts)
		shell.SSHBastionLogin(bastionClient, targetClient)
	} else {
		shell.SSHLogin(targetClient)
	}
}

func sendSSHKeys(target *api.Ec2Instance, bastion *api.Ec2Instance, ops *SshOps, sshOps *shell.SSHOpts) {

	ecc := api.NewEc2ConnClient()

	pl := api.Ec2ConnPayload{
		User: sshOps.User,
	}

	if ops.PubKey != "" {
		pl.PubKeyPath = ops.PubKey
	} else {
		hd, _ := homedir.Dir()
		pl.PubKeyPath = fmt.Sprintf("%v/.ssh/id_rsa.pub", hd)
	}

	if bastion != nil {
		pl.Instance = *bastion
		ecc.SendPublicKey(&pl)
	}

	if target != nil {
		pl.Instance = *target
		ecc.SendPublicKey(&pl)
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
