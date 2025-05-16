package sshconfig

import (
	"bufio"
	"bytes"
	"github.com/bmatcuk/doublestar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func parseContent(r io.Reader, cfg *Config, baseDir string, currentFile string) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		key := fields[0]
		args := fields[1:]

		switch key {
		case "Include":
			// 处理 Include 通配符
			for _, pattern := range args {
				// 支持 ~ 展开
				if strings.HasPrefix(pattern, "~") {
					if home, e := os.UserHomeDir(); e == nil {
						pattern = filepath.Join(home, pattern[1:])
					}
				}
				// 相对路径基于 baseDir
				if !filepath.IsAbs(pattern) {
					pattern = filepath.Join(baseDir, pattern)
				}
				matches, _ := doublestar.Glob(pattern)
				sort.Strings(matches)
				for _, m := range matches {
					sub, err := ioutil.ReadFile(m)
					if err != nil {
						continue
					}
					// 递归解析子文件
					parseContent(bytes.NewReader(sub), cfg, filepath.Dir(m), m)
				}
			}

		case "Host":
			// 新增 Host 块
			h := &Host{Hostnames: args,
				Source: currentFile}
			cfg.Hosts = append(cfg.Hosts, h)

		default:
			// 其他关键字放到最近的 Host 或全局
			param := &Param{Keyword: key, Args: args}
			if len(cfg.Hosts) == 0 {
				cfg.Globals = append(cfg.Globals, param)
			} else {
				last := cfg.Hosts[len(cfg.Hosts)-1]
				last.Params = append(last.Params, param)
			}
		}
	}
	return scanner.Err()
}
