package shell

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/compute"
	"github.com/mitchellh/go-homedir"
)

type SSHClient struct {
	Instance   compute.Ec2Instance
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

func NewSSHOpts() SSHOpts {
	o := SSHOpts{}
	o.IdentityFile = o.DefaultIdentityFile()
	o.User = os.Getenv("USER")
	o.Port = "22"
	o.Command = "login"
	o.AuthMethods = make([]ssh.AuthMethod, 0, 1)
	// s, _ := SSHKeyAuth()
	sock, err := SSHAgent()
	if err != nil {
		fmt.Println("SOCK ERR")
		spew.Dump(err)

		key, err := NewSSHKey(o.IdentityFile)
		keyAuth, err := SSHKeyAuth(key)
		if err != nil {
			fmt.Println("SSH Key Auth")
			spew.Dump(err)
		} else {

			o.AuthMethods = append(o.AuthMethods, keyAuth)
		}

	} else {
		o.AuthMethods = append(o.AuthMethods, sock)
	}
	return o
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

func NewSSHClient(inst compute.Ec2Instance, opts *SSHOpts) *SSHClient {

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
	fmt.Println("HERER I AM")
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

func (client *SSHClient) InteractiveSession() {

	session, err := client.SshClient.NewSession()

	if err != nil {
		fmt.Println("Cannot create SSH session ")
		fmt.Println(err)
		os.Exit(2)
	}

	client.SshSession = session

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
	// err = session.RequestPty("xterm-256color", termHeight, termWidth, termModes)
	fmt.Println("Height: ", termHeight, "TermWidth: ", termWidth)

	err = session.RequestPty("xterm", termHeight, termWidth, termModes)
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

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	session.Wait()
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

func AgentAuth() (ssh.AuthMethod, error) {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeysCallback(agent.NewClient(sock).Signers), nil
}

func SSHKeyAuth(priv *SSHKey) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey(priv.Body)
	if err != nil {
		fmt.Println("HERE IN KEY OPEN ERR")
		spew.Dump(err)
		if strings.Contains(err.Error(), "cannot decode encrypted private keys") {
			return decryptSSHKey(priv)
		} else if strings.Contains(err.Error(), "this private key is passphrase protected") {
			return decryptSSHKey(priv)
		}
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func decryptSSHKey(priv *SSHKey) (ssh.AuthMethod, error) {
	fmt.Fprintf(os.Stderr, "This SSH key is encrypted. Please enter passphrase for key '%s':", priv.Path)
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr)

	signer, err := _DecryptSSHKey(priv.Body, passphrase)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
func _DecryptSSHKey(key []byte, password []byte) (ssh.Signer, error) {
	block, _ := pem.Decode(key)
	pem, err := x509.DecryptPEMBlock(block, password)
	if err != nil {
		return nil, err
	}
	sshkey, err := x509.ParsePKCS1PrivateKey(pem)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(sshkey)
}

func (opts *SSHOpts) DefaultIdentityFile() string {

	hd, err := homedir.Dir()

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%v/.ssh/id_rsa", hd)

}

////////////////////
func (client *SSHClient) Login2(usePrivate bool) {

	ip := client.Instance.Ip

	if usePrivate {
		ip = client.Instance.PrivateIp
	}

	addr := net.JoinHostPort(ip, client.SSHOpts.Port)

	sshConn, err := ssh.Dial("tcp", addr, client.ClientConf)
	if err != nil {
		fmt.Println("Unable to dial into server")
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println("HERER I AM")
	client.SshClient = sshConn
	defer sshConn.Close()

	session, err := sshConn.NewSession()
	if err != nil {
		fmt.Printf("Cannot create SSH session to %v\n", addr)
		fmt.Println(err)
		os.Exit(2)
	}
	client.SshSession = session

	defer session.Close()

	termModes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// fileDescriptor := int(os.Stdin.Fd())
	fileDescriptor := int(os.Stdout.Fd())
	originalState, err := terminal.MakeRaw(fileDescriptor)

	defer terminal.Restore(fileDescriptor, originalState)
	termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
	if err != nil {
		log.Fatal(err)
	}
	// err = session.RequestPty("xterm-256color", termHeight, termWidth, termModes)
	fmt.Println("Height: ", termHeight, "TermWidth: ", termWidth)

	err = session.RequestPty("xterm", termHeight, termWidth, termModes)
	// err = session.RequestPty("xterm-256color", 100, 100, termModes)
	if err != nil {
		fmt.Println("Error requesting a terminal")
		os.Exit(2)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	if client.SSHOpts.ForwardAuthSock {
		client.ForwardAuthSock()
	}

	err = session.Shell()

	if err != nil {
		fmt.Println("Error starting shell")
		os.Exit(2)
	}

	resizeTerminal(session, 10)
	session.Wait()

}

func resizeTerminal(aSession *ssh.Session, timeout int) {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	c := 0
	go func() {
		for {
			select {
			case <-ticker.C:
				c++
				// do stuff
				aFileDes := int(os.Stdout.Fd())
				termWidth, termHeight, _ := terminal.GetSize(aFileDes)

				// fmt.Println(termWidth, termHeight, "\n")
				aSession.WindowChange(termHeight, termWidth)
				if c >= timeout {
					// aSession.Close()
				}
				// buffer := bytes.Buffer{}
				// buffer.Write([]byte("tput cols && tput lines\n"))
				// buffer.Write([]byte("uptime\n"))
				// var br = []byte("uptime\n")
				// res, err := aSession.SendRequest("uptime\n", true, br)
				// spew.Dump(res)
				// if err != nil {
				// 	fmt.Println(err)
				// }

				// os.Stdin = &buffer
				// aSession.Stdin = &buffer
				// aSession.Stdin = os.Stdin
				// log.Printf("Resize: %dx%d", termHeight, termWidth)
				// c := fmt.Sprintf("resizecons %dx%d", termWidth, termHeight)
				// aSession.Run(c)
				// aSession.RequestPty("xterm", termHeight, termWidth, termModes)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
