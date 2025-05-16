package sshconfig

type Config struct {
	Globals []*Param
	Hosts   []*Host
}

type Host struct {
	Hostnames []string
	Params    []*Param
	Source    string
}

type Param struct {
	Keyword string   // 如 "HostName", "User"
	Args    []string // 关键字后的所有字段
}

const (
	HostKeyword     = "Host"
	HostNameKeyword = "HostName"
	PortKeyword     = "Port"
	UserKeyword     = "User"
)

// GetParam 返回第一个匹配 keyword 的 Param，找不到返回 nil
func (h *Host) GetParam(keyword string) *Param {
	for _, p := range h.Params {
		if p.Keyword == keyword {
			return p
		}
	}
	return nil
}

// Value 返回第一个 arg 或空串
func (p *Param) Value() string {
	if len(p.Args) > 0 {
		return p.Args[0]
	}
	return ""
}
