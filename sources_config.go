//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-12-11
//

package ssh

import (
	"fmt"
	"github.com/zimmortal/alfred-ssh/internal/sshconfig"
	"log"
	"net/url"
	"strconv"
	"strings"
)

// ConfigHost is Host parsed from SSH config-format files.
type ConfigHost struct {
	BaseHost
	forcePort     bool
	forceUsername bool
}

// UID implements Host.
func (h *ConfigHost) UID() string { return UIDForHost(h) }

// SetPort implements Host.
func (h *ConfigHost) SetPort(i int) {
	h.port = i
	h.forcePort = true
}

// SetUsername implemeents Host.
func (h *ConfigHost) SetUsername(n string) {
	h.username = n
	h.forceUsername = true
}

// SSHURL returns a URL based on the Host value from the config file,
// *not* the Hostname.
func (h *ConfigHost) SSHURL() *url.URL {
	u := &url.URL{
		Scheme: "ssh",
		Host:   h.Name(),
	}
	if h.forcePort {
		u.Host = fmt.Sprintf("%s:%d", u.Host, h.Port())
	}
	if h.forceUsername {
		u.User = url.User(h.Username())
	}
	return u
}

// MoshCmd implements Host.
func (h *ConfigHost) MoshCmd(path string) string {
	if path == "" {
		path = "mosh"
	}
	cmd := path + " "
	if h.forcePort {
		cmd += fmt.Sprintf("--ssh 'ssh -p %d' ", h.Port())
	}
	if h.forceUsername && h.Username() != "" {
		cmd += h.Username() + "@"
	}
	cmd += h.Name()
	return cmd
}

// ConfigSource implements Source for ssh config-formatted files.
type ConfigSource struct {
	baseSource
}

// NewConfigSource creates a new ConfigSource from an ssh configuration file.
func NewConfigSource(path, name string, priority int) *ConfigSource {
	s := &ConfigSource{}
	s.Filepath = path
	s.name = name
	s.priority = priority
	return s
}

// Hosts implements Source.
func (s *ConfigSource) Hosts() []Host {
	if s.hosts == nil {
		hosts := parseConfigFile(s.Filepath)
		log.Printf("[source/load/config] %d host(s) in '%s'", len(hosts), s.Name())
		s.hosts = make([]Host, len(hosts))
		for i, h := range hosts {
			h.source = s.Name()
			s.hosts[i] = Host(h)
		}
	}
	return s.hosts
}

// parseConfigFile parse an SSH config file.
func parseConfigFile(path string) []*ConfigHost {
	var hosts []*ConfigHost

	// 直接用你在 internal/sshconfig 定义的 ParseFile
	cfg, err := sshconfig.ParseFile(path)
	if err != nil {
		log.Printf("[config/%s] Parse error: %s", path, err)
		return hosts
	}

	for _, e := range cfg.Hosts {
		var (
			p    *sshconfig.Param
			port = 22
			hn   string // hostname
			user string
		)

		// 原来的字段提取逻辑不变
		p = e.GetParam(sshconfig.HostKeyword)
		if p != nil {
			hn = p.Value()
		}

		p = e.GetParam(sshconfig.HostNameKeyword)
		if p != nil {
			hn = p.Value()
		}

		p = e.GetParam(sshconfig.PortKeyword)
		if p != nil {
			if i, err := strconv.Atoi(p.Value()); err == nil {
				port = i
			} else {
				log.Printf("Bad port in %s: %s", path, err)
			}
		}

		p = e.GetParam(sshconfig.UserKeyword)
		if p != nil {
			user = p.Value()
		}

		for _, n := range e.Hostnames {
			if strings.ContainsAny(n, "*!?") {
				continue
			}

			h := &ConfigHost{
				BaseHost: BaseHost{
					name:     n,
					hostname: n,
					port:     port,
					username: user,
				},
			}
			if hn != "" {
				h.hostname = hn
			}
			hosts = append(hosts, h)
		}
	}
	return hosts
}
