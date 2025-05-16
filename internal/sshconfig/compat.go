// internal/sshconfig/compat.go
package sshconfig

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// internal/sshconfig/sshconfig.go

// ParseFile 解析指定路径的 SSH 配置，支持 Include
func ParseFile(path string) (*Config, error) {
	// 1. 读入文件内容
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// 2. 准备 Config
	cfg := &Config{}
	// 3. 计算初始 baseDir 和 currentFile（取绝对路径）
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	baseDir := filepath.Dir(abs)
	// 4. 调用核心解析，传入当前文件名
	if err := parseContent(bytes.NewReader(data), cfg, baseDir, abs); err != nil {
		return nil, err
	}
	return cfg, nil
}

// 兼容旧接口：保留 Parse(r io.Reader)，但它只是调用 ParseFile
func Parse(r io.Reader) (*Config, error) {
	// 既然不知道真正路径，就退回到 ParseFile("~/ .ssh/config")
	home, _ := os.UserHomeDir()
	mainPath := filepath.Join(home, ".ssh", "config")
	return ParseFile(mainPath)
}
