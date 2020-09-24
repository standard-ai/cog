package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const sshAgentLifetimeSecs = 86400

type Key struct {
	prv  interface{}
	pub  []byte
	cert *ssh.Certificate
}

func NewKey() (*Key, error) {
	if verbose > 0 {
		log.Info("Generating a new SSH key")
	}
	prv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	pub, err := ssh.NewPublicKey(&prv.PublicKey)
	if err != nil {
		return nil, err
	}
	return &Key{prv, ssh.MarshalAuthorizedKey(pub), nil}, nil
}

func getSock() (sshAuthSock string, cleanup func()) {
	sshAuthSock = viper.GetString("ssh_identity_agent")
	if sshAuthSock != "" {
		if verbose > 0 {
			log.Infof("Using %s as SSH_AUTH_SOCK", sshAuthSock)
		}
	} else {
		sshAuthSock = os.Getenv("SSH_AUTH_SOCK")
	}
	if sshAuthSock != "" {
		if _, err := os.Stat(sshAuthSock); err != nil {
			sshAuthSock = ""
		} else if verbose > 0 {
			log.Info("Using existing SSH agent")
		}
	}
	if sshAuthSock == "" {
		sshAuthSock, cleanup = startAgent()
		os.Setenv("SSH_AUTH_SOCK", sshAuthSock)
	}
	return
}

func addCertToAgent(k *Key, socket string) error {
	if verbose > 0 {
		log.Info("Adding certificate to SSH agent")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("Failed to open SSH_AUTH_SOCK: %v", err)
	}

	agentClient := agent.NewClient(conn)
	if err = agentClient.Add(agent.AddedKey{
		PrivateKey:   k.prv,
		Certificate:  k.cert,
		LifetimeSecs: sshAgentLifetimeSecs,
	}); err != nil {
		log.Fatalf("Adding cert: %v", err)
	}
	return nil
}

func removeCertFromAgent(k *Key, socket string) error {
	if verbose > 0 {
		log.Info("Removing certificate from SSH agent")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("Failed to open SSH_AUTH_SOCK: %v", err)
	}

	agentClient := agent.NewClient(conn)
	return agentClient.Remove(k.cert)
}

func startAgent() (sshAuthSock string, cleanup func()) {
	if verbose > 0 {
		log.Info("Starting SSH agent")
	}

	bin, err := exec.LookPath("ssh-agent")
	if err != nil {
		log.Fatal("Could not find ssh-agent")
	}

	cmd := exec.Command(bin, "-s")
	cmd.Env = []string{} // do not let the users env influence ssh-agent behavior
	cmd.Stderr = new(bytes.Buffer)
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("%s failed: %v\n%s", strings.Join(cmd.Args, " "), err, cmd.Stderr)
	}
	// Output looks like:
	//	SSH_AUTH_SOCK=/tmp/ssh-P65gpcqArqvH/agent.15541; export SSH_AUTH_SOCK;
	//	SSH_AGENT_PID=15542; export SSH_AGENT_PID;
	//	echo Agent pid 15542;
	fields := bytes.Split(out, []byte(";"))
	line := bytes.SplitN(fields[0], []byte("="), 2)
	if string(line[0]) != "SSH_AUTH_SOCK" {
		log.Fatalf("Could not find key SSH_AUTH_SOCK in %q", fields[0])
	}
	sshAuthSock = string(line[1])

	line = bytes.SplitN(fields[2], []byte("="), 2)
	line[0] = bytes.TrimLeft(line[0], "\n")
	if string(line[0]) != "SSH_AGENT_PID" {
		log.Fatalf("Could not find key SSH_AGENT_PID in %q", fields[0])
	}
	pidStr := string(line[1])
	sshAgentPid, err := strconv.Atoi(pidStr)
	if err != nil {
		log.Fatalf("Atoi(%q): %v", pidStr, err)
	}

	return sshAuthSock, func() {
		proc, _ := os.FindProcess(sshAgentPid)
		if proc != nil {
			proc.Kill()
		}
		os.RemoveAll(filepath.Dir(sshAuthSock))
	}
}
