package old

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/api"
)

type SSHSessionOpts struct {
	Tty      bool
	Username string
	Command  string
	AuthSock bool
	Identity string
}

type SSHSession struct {
	Instance api.Ec2Instance
	Bastion  api.Ec2Instance
	Opts     SSHSessionOpts
}

func NewSSH(inst api.Ec2Instance, bastion api.Ec2Instance, opts SSHSessionOpts) *SSHSession {
	new := &SSHSession{
		Instance: inst,
		Bastion:  bastion,
		Opts:     opts,
	}

	return new
}

func (sess *SSHSession) buildArgs(args []string) []string {

	if sess.Opts.Tty {
		args = append(args, "-t")
	}

	if sess.Opts.AuthSock {
		args = append(args, "-A")
	}

	return args
}

func (sess *SSHSession) Login() {

	cmd := "ssh"

	var args []string

	// args = append(args, "-vvv -i /home/jhardy/.ssh/id_rsa")

	if len(sess.Bastion.Ip) > 0 {
		args = sess.bastionArgs(args)
		fmt.Println("USRING BASTION")
	}

	// args = sess.buildArgs(args)

	args = sess.serverArgs(args)

	if len(sess.Opts.Command) > 0 {
		args = append(args, sess.Opts.Command)
	}
	spew.Dump(args)
	ssh := exec.Command(cmd, args...)
	ssh.Env = append(os.Environ(), "AWSSH=1")
	spew.Dump(ssh)
	ssh.Stderr = os.Stderr
	ssh.Stdout = os.Stdout
	ssh.Stdin = os.Stdin

	ssh.Run()
}

func (sess *SSHSession) Send() {

	var args []string

	args = sess.buildArgs(args)

	args = sess.serverArgs(args)

	if len(sess.Opts.Command) > 0 {
		args = append(args, sess.Opts.Command)
	}

	spew.Dump(args)

	out, _ := exec.Command("ssh", args...).Output()

	spew.Dump(out)
	fmt.Println(string(out))
}

func (sess *SSHSession) serverArgs(args []string) []string {

	m := make(map[string]string)

	m["user"] = sess.Opts.Username

	if len(sess.Bastion.Ip) > 0 {
		m["ip"] = sess.Instance.PrivateIp
	} else {
		m["ip"] = sess.Instance.Ip
	}

	tmpl, _ := template.New("SSHLogin").Parse("{{ .user }}@{{ .ip }}")

	var buf bytes.Buffer

	tmpl.Execute(&buf, m)

	args = append(args, buf.String())

	return args

}

func (sess *SSHSession) bastionArgs(args []string) []string {

	m := make(map[string]string)

	m["user"] = sess.Opts.Username
	m["ip"] = sess.Bastion.Ip

	tmpl, _ := template.New("Bastion").Parse("ProxyCommand=\"ssh -W %h:%p {{ .user }}@{{ .ip }}\"")

	var buf bytes.Buffer

	tmpl.Execute(&buf, m)

	args = append(args, "-o")
	args = append(args, buf.String())

	return args
}
