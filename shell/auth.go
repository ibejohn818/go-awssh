package shell

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

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
		if strings.Contains(err.Error(), "cannot decode encrypted private keys") {
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
