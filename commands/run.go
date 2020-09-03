package commands

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/ibejohn818/go-awssh/api"
	"github.com/ibejohn818/go-awssh/config"
	"github.com/ibejohn818/go-awssh/shell"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func AddRunCmd(cmd *cobra.Command, conf *config.AwsshConf) {

	sshOpts := &shell.SSHOpts{}
	confirm := false
	useBastion := false

	runCmd := &cobra.Command{
		Use:   "run [Instance Name Contains]",
		Args:  cobra.ArbitraryArgs,
		Short: "Send commands: to servers VIA SSH",
		Long: `Send commands to multiple servers and through a bastion if needed (-b).
Use strings argument to filter servers by their name tag. Strings starting with ^ will denote a negative match. 
EXAMPLES:
awssh run web ^java -c "uptime" 

The above will run the "uptime" command on servers that contain "web" and do not container "java"
in their name tag.`,
		Run: func(cmd *cobra.Command, args []string) {

			if len(sshOpts.Command) <= 0 {
				fmt.Println("--cmd/-c Command is required!")
				os.Exit(2)
			}

			ec2Coll := api.GetServers(conf)

			ec2Res := ec2Coll.Filtered(args)

			if len(ec2Res) <= 0 {
				fmt.Println("No instances in scope")
				os.Exit(0)
			}

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

			var wg sync.WaitGroup

			if useBastion {

				bastion := selectInstance(ec2Coll.Instances, "Select bastion instance")
				bSSH := shell.NewSSHClient(bastion, sshOpts)
				bConn := shell.BastionRunClient(bSSH)
				fmt.Println("------------------")
				fmt.Printf("Command: %v \n", sshOpts.Command)
				fmt.Println("Bastion:")
				fmt.Println(bastion.GetLine())
				fmt.Println("Instances:")

				for _, v := range ec2Res {
					fmt.Println(v.GetLinePrivate())
				}
				fmt.Println("------------------")

				if !confirm {
					confirmMsg()
				}

				for k, _ := range ec2Res {

					wg.Add(1)
					go func(bClient *ssh.Client, inst api.Ec2Instance, opts *shell.SSHOpts) {
						sClient := shell.NewSSHClient(inst, opts)
						ip := inst.PrivateIp
						addr := net.JoinHostPort(ip, opts.Port)
						sConn, err := bClient.Dial("tcp", addr)
						if err != nil {
							fmt.Println("Server client dial error")
						}
						ncc, chans, reqs, err := ssh.NewClientConn(sConn, addr, sClient.ClientConf)
						if err != nil {
							fmt.Println("Server new client error")
						}
						sshConn := ssh.NewClient(ncc, chans, reqs)
						sClient.SshClient = sshConn

						session, err := sshConn.NewSession()
						if err != nil {
							fmt.Println("Session connect error")
						}
						sClient.SshSession = session

						sClient.ForwardAuthSock()

						defer session.Close()

						var buffOut bytes.Buffer
						session.Stdout = &buffOut
						session.Run(opts.Command)
						fmt.Println(inst.GetLinePrivate())
						fmt.Println("------------------")
						fmt.Println(buffOut.String())
						fmt.Println("------------------")

						wg.Done()

					}(bConn, ec2Res[k], sshOpts)

				}
			} else {
				fmt.Printf("Command: %v \n", sshOpts.Command)

				fmt.Println("Instances:")

				for _, v := range ec2Res {
					fmt.Println(v.GetLine())
				}
				fmt.Println("------------------")

				if !confirm {
					confirmMsg()
				}
				for k, _ := range ec2Res {

					wg.Add(1)
					go func(inst api.Ec2Instance, opts *shell.SSHOpts) {

						sshClient := shell.NewSSHClient(inst, opts)
						ip := inst.Ip

						addr := net.JoinHostPort(ip, sshClient.SSHOpts.Port)

						sshConn, err := ssh.Dial("tcp", addr, sshClient.ClientConf)
						if err != nil {
							fmt.Println("Unable to dial into server")
							fmt.Println(err)
							return
						}

						sshClient.SshClient = sshConn

						session, err := sshConn.NewSession()
						if err != nil {
							fmt.Println("Session connect error")
							return
						}

						sshClient.SshSession = session

						sshClient.ForwardAuthSock()

						var buffOut bytes.Buffer
						session.Stdout = &buffOut
						session.Run(opts.Command)
						fmt.Println(inst.GetLine())
						fmt.Println("------------------")
						fmt.Println(buffOut.String())
						fmt.Println("------------------")
						wg.Done()
					}(ec2Res[k], sshOpts)
				}

			}

			wg.Wait()

		},
	}

	flags := runCmd.Flags()

	flags.StringVarP(&sshOpts.User, "username", "u", "ec2-user", "SSH Username")
	conf.VpConf.BindPFlag("username", flags.Lookup("username"))

	flags.BoolVarP(&useBastion, "bastion", "b", false, "Connect via bastion host")
	flags.BoolVarP(&confirm, "yes", "y", false, "Skip confirmation")
	flags.BoolVarP(&sshOpts.ForwardAuthSock, "auth", "A", false, "Pass SSH AUTH SOCK")
	flags.StringVarP(&sshOpts.Port, "port", "p", "22", "SSH Port")
	flags.StringVarP(&sshOpts.IdentityFile, "identity", "i", "", "Path to ssh key file")
	flags.StringVarP(&sshOpts.Command, "cmd", "c", "", "*REQUIRED: Command to run")

	cmd.AddCommand(runCmd)
}

func confirmMsg() {

	prompt := promptui.Prompt{
		Label:     "Run commands",
		IsConfirm: true,
	}

	retVal, _ := prompt.Run()

	if retVal != "y" {
		os.Exit(0)
	}
	fmt.Println("------------------")
	fmt.Println("Results:")
	fmt.Println("------------------")
}
