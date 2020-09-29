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
	"github.com/ibejohn818/awssh/api"
	"github.com/mitchellh/go-homedir"
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
	ForcePrivKey    bool
	UsePrivateIp    bool
}

type SSHKey struct {
	Path string
	Body []byte
}

func NewSSHOpts(ops ...func(*SSHOpts)) SSHOpts {
	o := SSHOpts{}
	for _, v := range ops {
		v(&o)
	}

	if len(o.User) <= 0 {

		o.User = os.Getenv("USER")
	}

	if len(o.Port) <= 0 {
		o.Port = "22"
	}

	o.Command = "login"

	o.AuthMethods = make([]ssh.AuthMethod, 0, 0)

	if len(o.IdentityFile) <= 0 {
		o.AuthMethods = append(o.AuthMethods, appendSocketAuth())
	}

	privKeyChk1 := len(o.IdentityFile) <= 0 && len(o.AuthMethods) <= 0
	privKeyChk2 := len(o.IdentityFile) <= 0 && o.ForcePrivKey
	//if len(o.IdentityFile) <= 0 && len(o.AuthMethods) <= 0 {
	if privKeyChk1 || privKeyChk2 {
		o.IdentityFile = o.DefaultIdentityFile()
	}

	if len(o.IdentityFile) > 0 {
		o.AuthMethods = append(o.AuthMethods,
			appendPublicKeyAuth(o.IdentityFile))
	}

	return o
}

func appendSocketAuth() ssh.AuthMethod {
	// s, _ := SSHKeyAuth()
	// sock, err := SSHAgent()
	sock, err := AgentAuth()
	if err != nil {
		fmt.Println("SOCK ERR")
		spew.Dump(err)

	}
	return sock

}

func appendPublicKeyAuth(aFile string) ssh.AuthMethod {

	key, err := NewSSHKey(aFile)
	keyAuth, err := SSHKeyAuth(key)

	if err != nil {
		fmt.Println("SSH Key Auth")
		spew.Dump(err)
	}
	return keyAuth

}

func NewSSHKey(path string) (*SSHKey, error) {

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key := SSHKey{
		Path: path,
		Body: bytes,
	}

	return &key, nil
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

func (client *SSHClient) InteractiveSession() {

	session, err := client.SshClient.NewSession()

	defer client.SshClient.Close()

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

	xterm := "xterm-256color"
	err = session.RequestPty(xterm, termHeight, termWidth, termModes)

	if err != nil {
		fmt.Println("Error requesting a terminal")
		os.Exit(2)
	}

	if client.SSHOpts.ForwardAuthSock {
		client.ForwardAuthSock()
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	err = session.Shell()

	resizeTerminal(session)

	if err != nil {
		fmt.Println("Error starting shell")
		os.Exit(2)
	}

	session.Wait()
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

	client.InteractiveSession()
}

func SSHLogin(client *SSHClient) {

	ip := client.Instance.Ip

	if client.SSHOpts.UsePrivateIp {
		ip = client.Instance.PrivateIp
	}

	addr := net.JoinHostPort(ip, client.SSHOpts.Port)

	sshConn, err := ssh.Dial("tcp", addr, client.ClientConf)
	if err != nil {
		fmt.Println("Unable to dial into server")
		fmt.Println(err)
		os.Exit(2)
	}

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
	signer := agent.NewClient(sock).Signers

	return ssh.PublicKeysCallback(signer), nil
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

func (opts *SSHOpts) getConnecdtionIp() string {
	return ""
}

func resizeTerminal(aSession *ssh.Session) {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				aFileDes := int(os.Stdout.Fd())
				termWidth, termHeight, _ := terminal.GetSize(aFileDes)

				// fmt.Println(termWidth, termHeight, "\n")
				aSession.WindowChange(termHeight, termWidth)

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
