package shell

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/go-awssh/api"
	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

type SSHClient struct {
	Instance   api.Ec2Instance
	ClientConf *ssh.ClientConfig
	SSHOpts    *SSHOpts
	SshClient  *ssh.Client
	SshSession *ssh.Session
}

type SSHOpts struct {
	ForwardAuthSock bool
	Tty             bool
	IdentityFile    string
	User, Port      string
	Command         string
	AuthMethods     []ssh.AuthMethod
}

type SSHKey struct {
	Path string
	Body []byte
}

func NewSSHKey(path string) (*SSHKey, error) {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	new := &SSHKey{
		Path: path,
		Body: bytes,
	}

	return new, nil
}

func NewSSHClient(inst api.Ec2Instance, opts *SSHOpts) *SSHClient {

	conf := &ssh.ClientConfig{
		User:            opts.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            opts.AuthMethods,
	}
	sess := &SSHClient{
		Instance:   inst,
		SSHOpts:    opts,
		ClientConf: conf,
	}

	return sess
}

func BastionRunClient(bClient *SSHClient) *ssh.Client {
	ip := bClient.Instance.Ip
	bAddr := net.JoinHostPort(ip, bClient.SSHOpts.Port)

	bConn, err := ssh.Dial("tcp", bAddr, bClient.ClientConf)
	if err != nil {
		log.Fatal(err)
	}

	return bConn
}

func SSHBastionLogin(bClient *SSHClient, client *SSHClient) {

	ip := bClient.Instance.Ip
	bAddr := net.JoinHostPort(ip, client.SSHOpts.Port)

	pip := client.Instance.PrivateIp
	addr := net.JoinHostPort(pip, client.SSHOpts.Port)

	bConn, err := ssh.Dial("tcp", bAddr, bClient.ClientConf)
	if err != nil {
		log.Fatal(err)
	}

	sConn, err := bConn.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(sConn, addr, client.ClientConf)

	sshConn := ssh.NewClient(ncc, chans, reqs)
	client.SshClient = sshConn
	defer sshConn.Close()

	client.InteractiveSession()
}

func SSHLogin(client *SSHClient) {

	ip := client.Instance.Ip

	addr := net.JoinHostPort(ip, client.SSHOpts.Port)

	sshConn, err := ssh.Dial("tcp", addr, client.ClientConf)
	if err != nil {
		fmt.Println("Unable to dial into server")
		fmt.Println(err)
		os.Exit(2)
	}
	defer sshConn.Close()

	client.SshClient = sshConn

	client.InteractiveSession()
}

func (client *SSHClient) InteractiveSession() {

	session, err := client.SshClient.NewSession()

	if err != nil {
		fmt.Println("Cannot create SSH session ")
		fmt.Println(err)
		os.Exit(2)
	}

	client.SshSession = session

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	defer session.Close()

	termModes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fileDescriptor := int(os.Stdin.Fd())
	originalState, err := terminal.MakeRaw(fileDescriptor)

	defer terminal.Restore(fileDescriptor, originalState)
	termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
	if err != nil {
		log.Fatal(err)
	}
	err = session.RequestPty("xterm-256color", termHeight, termWidth, termModes)
	if err != nil {
		fmt.Println("Error requesting a terminal")
		os.Exit(2)
	}

	if client.SSHOpts.ForwardAuthSock {
		client.ForwardAuthSock()
	}

	if len(client.SSHOpts.Command) > 0 {
		session.Run(client.SSHOpts.Command)
	}

	err = session.Shell()
	if err != nil {
		fmt.Println("Error starting shell")
		os.Exit(2)
	}

	session.Wait()
}

func (client *SSHClient) Login(usePrivate bool) {

	ip := client.Instance.Ip

	if usePrivate {
		ip = client.Instance.PrivateIp
	}
	spew.Dump(client.Instance)
	addr := net.JoinHostPort(ip, client.SSHOpts.Port)

	sshConn, err := ssh.Dial("tcp", addr, client.ClientConf)
	if err != nil {
		fmt.Println("Unable to dial into server")
		fmt.Println(err)
		os.Exit(2)
	}
	client.SshClient = sshConn
	defer sshConn.Close()

	session, err := sshConn.NewSession()
	if err != nil {
		fmt.Printf("Cannot create SSH session to %v\n", addr)
		fmt.Println(err)
		os.Exit(2)
	}

	client.SshSession = session

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	defer session.Close()

	termModes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fileDescriptor := int(os.Stdin.Fd())
	originalState, err := terminal.MakeRaw(fileDescriptor)

	defer terminal.Restore(fileDescriptor, originalState)

	err = session.RequestPty("xterm-256color", 100, 100, termModes)
	if err != nil {
		fmt.Println("Error requesting a terminal")
		os.Exit(2)
	}

	if client.SSHOpts.ForwardAuthSock {
		client.ForwardAuthSock()
	}

	err = session.Shell()
	if err != nil {
		fmt.Println("Error starting shell")
		os.Exit(2)
	}

	session.Wait()

}

func SSHRunCommand(client *SSHClient) {

}

func (client *SSHClient) ForwardAuthSock() {
	err := agent.ForwardToRemote(client.SshClient, os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		fmt.Println("ERR remote agent")
		spew.Dump(err)
	}
	err = agent.RequestAgentForwarding(client.SshSession)
	if err != nil {
		fmt.Println("Request agent forwaridng errro")
		spew.Dump(err)
	}
}

func SSHAgent() (ssh.AuthMethod, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeysCallback(agent.NewClient(sock).Signers), nil
}

func (opts *SSHOpts) DefaultIdentityFile() string {

	hd, err := homedir.Dir()

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%v/.ssh/id_rsa", hd)

}
