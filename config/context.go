package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

// Context structure
type Context struct {
	config  *Config
	context *ContextParams
}

// NewContext with received params and uninitialized configuration
func NewContext(params ContextParams) Context {
	return Context{
		context: &ContextParams{},
		config:  &Config{},
	}
}

// ContextParams is the set of environment configurations
type ContextParams struct {
	Remote               string
	Infrastructure       string
	InfrastructureDomain string
	ServiceDomain        string
	Username             string
	Token                string
}

// Remote used on the context
func (c *Context) Remote() string {
	return c.context.Remote
}

// Infrastructure used on the context
func (c *Context) Infrastructure() string {
	infra := c.context.Infrastructure

	if infra == "localhost" || strings.HasPrefix(infra, "localhost:") {
		return "http://" + infra
	}

	return infra
}

// InfrastructureDomain used on the context
func (c *Context) InfrastructureDomain() string {
	return c.context.InfrastructureDomain
}

// ServiceDomain used on the context
func (c *Context) ServiceDomain() string {
	return c.context.ServiceDomain
}

// Username used on the context
func (c *Context) Username() string {
	return c.context.Username
}

// Token used on the context
func (c *Context) Token() string {
	return c.context.Token
}

// Config gets the configuration
func (c *Context) Config() *Config {
	return c.config
}

// SetEndpoint for the context
func (c *Context) SetEndpoint(remote string) error {
	var conf = c.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	if !rl.Has(remote) {
		return fmt.Errorf(`error loading selected remote "%v"`, remote)
	}

	var r = rl.Get(remote)

	c.context.Remote = remote
	c.context.Infrastructure = r.InfrastructureServer()
	c.context.InfrastructureDomain = getRemoteAddress(r.Infrastructure)
	c.context.ServiceDomain = r.Service
	c.context.Username = r.Username
	c.context.Token = r.Token
	return nil
}

// Setup the environment
func Setup(path string) (wectx Context, err error) {
	wectx = NewContext(ContextParams{})
	path, err = filepath.Abs(path)

	if err != nil {
		return wectx, err
	}

	var c = &Config{
		Path: path,
	}

	if err = c.Load(); err != nil {
		return wectx, err
	}

	wectx.config = c
	return wectx, nil
}

func getRemoteAddress(address string) string {
	if strings.HasPrefix(address, "https://api.") {
		address = strings.TrimPrefix(address, "https://api.")
	}

	var h, _, err = net.SplitHostPort(address)

	if err != nil {
		return address
	}

	return h
}
