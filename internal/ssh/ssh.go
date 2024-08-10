package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/amoydavid/godeployer/internal/config"
	"golang.org/x/crypto/ssh"
)

// Client 表示一个SSH客户端连接
type Client struct {
	*ssh.Client
	config *config.Config
}

// NewClient 创建一个新的SSH客户端
func NewClient(cfg *config.Config) (*Client, error) {
	key, err := ioutil.ReadFile(cfg.SSHKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: cfg.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(cfg.Host, cfg.SSHPort), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &Client{
		Client: client,
		config: cfg,
	}, nil
}

// Run 在远程主机上执行命令
func (c *Client) Run(command string) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %v", err)
	}

	return string(output), nil
}

// Close 关闭SSH连接
func (c *Client) Close() error {
	return c.Client.Close()
}

func (c *Client) RunCommand(cmd string) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	session.Stdout = &stdout

	err = session.Run(cmd)
	return stdout.String(), err
}
