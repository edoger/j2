// Copyright 2021 Qingshan Luo <edoger@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

var Cfg = NewConfig()

type Config struct {
	PageSize   int       `yaml:"pageSize"`   // 每页显示多少个服务器（默认10）
	SortBy     string    `yaml:"sortBy"`     // 排序方式（name, host, disable）
	AutoClear  bool      `yaml:"autoClear"`  // 自动清屏
	PrivateKey string    `yaml:"privateKey"` // 全局的私钥路径（可被服务器设置覆盖）
	Password   string    `yaml:"password"`   // 全局的登录密码（可被服务器设置覆盖）
	Servers    []*Server `yaml:"servers"`    // 远程服务器列表

	Auth  ssh.AuthMethod `yaml:"-"`
	Page  int            `yaml:"-"`
	Group string         `yaml:"-"`
}

func NewConfig() *Config {
	return &Config{Page: 1}
}

type Server struct {
	Name       string `yaml:"name"`       // 名称（可被用于搜索和快速连接）
	User       string `yaml:"user"`       // 登录用户名
	Host       string `yaml:"host"`       // 登录主机名或IP地址
	Port       int    `yaml:"port"`       // 登录端口（默认22）
	PrivateKey string `yaml:"privateKey"` // 私钥路径（为空时不适用）
	Password   string `yaml:"password"`   // 登录密码（私钥登录优先，没有私钥则使用密码）
	Desc       string `yaml:"desc"`       // 简短的描述
	Group      string `yaml:"group"`      // 分组

	Auth ssh.AuthMethod `yaml:"-"`
	Addr string         `yaml:"-"`
}

func (c *Config) Init() error {
	if s := os.Getenv("J2_CONFIG_FILE"); s != "" {
		return c.from(s)
	}
	if s := filepath.Join(os.Getenv("HOME"), ".j2.yaml"); c.exist(s) {
		return c.from(s)
	}
	return c.from(".j2.yaml")
}

func (c *Config) NextPage() {
	if c.Page*c.PageSize >= len(c.AllList()) {
		c.Page = 1
	} else {
		c.Page++
	}
}

func (c *Config) PrevPage() {
	c.Page--
	if c.Page < 1 {
		if n := len(c.AllList()); n > 0 {
			c.Page = (n - n%c.PageSize) / c.PageSize
			if n%c.PageSize > 0 {
				c.Page++
			}
		} else {
			c.Page = 1
		}
	}
}

func (c *Config) AllList() []*Server {
	if c.Group == "" || c.Group == "default" {
		return c.Servers
	}
	r := make([]*Server, 0)
	for i, j := 0, len(c.Servers); i < j; i++ {
		if c.Servers[i].Group == c.Group {
			r = append(r, c.Servers[i])
		}
	}
	return r
}

func (c *Config) PageList() []*Server {
	var list []*Server
	all := c.AllList()
	if n := len(all); n > 0 {
		begin := (c.Page - 1) * c.PageSize
		end := c.Page * c.PageSize
		if begin >= n {
			begin = 0
			end = c.PageSize
		}
		if end >= n {
			list = all[begin:]
		} else {
			list = all[begin:end]
		}
	}
	return list
}

func (c *Config) Summary(list []*Server) []string {
	if len(list) == 0 {
		return nil
	}
	counts := []int{4, 4, 4, 5, 4} // name user host group desc
	for i, j := 0, len(list); i < j; i++ {
		if n := runewidth.StringWidth(list[i].Name); n > counts[0] { // name
			counts[0] = n
		}
		if n := runewidth.StringWidth(list[i].User); n > counts[1] { // user
			counts[1] = n
		}
		if n := runewidth.StringWidth(list[i].Host); n > counts[2] { // host
			counts[2] = n
		}
		if n := runewidth.StringWidth(list[i].Group); n > counts[3] { // group
			counts[3] = n
		}
		if n := runewidth.StringWidth(list[i].Desc); n > counts[4] { // desc
			counts[4] = n
		}
	}
	for i, j := 0, len(counts); i < j; i++ {
		if counts[i] == 0 {
			counts[i] = 2
		}
	}
	num := len(strconv.Itoa(len(list)))
	format := fmt.Sprintf(" %%-%ds  %%-%ds  %%-%ds %%-%ds  %%-%ds  %%-%ds", num, counts[0], counts[1], counts[2], counts[3], counts[4])
	prefix := color.New(color.FgHiGreen).Sprint(" **")
	summary := make([]string, 0, len(list)+1)
	summary = append(summary, "   "+color.New(color.FgYellow).Sprintf(format, "", "NAME", "USER", "HOST", "GROUP", "DESC"))
	for i, j := 0, len(list); i < j; i++ {
		// name user host group desc
		args := []interface{}{
			strconv.Itoa(i + 1),
			c.stuff(list[i].Name), c.stuff(list[i].User), c.stuff(list[i].Host),
			c.stuff(list[i].Group), c.stuff(list[i].Desc),
		}
		summary = append(summary, prefix+color.New(color.FgCyan).Sprintf(format, args...))
	}
	return summary
}

func (c *Config) stuff(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func (c *Config) ShowSummary() {
	ClearScreen()
	ShowTitle()

	var n int
	summary := c.Summary(c.PageList())
	for i, j := 0, len(summary); i < j; i++ {
		if nn := runewidth.StringWidth(summary[i]); nn > n {
			n = nn
		}
	}
	if n == 0 {
		n = 55
	}
	line := color.RedString(strings.Repeat("-", n))

	Echo(line)
	if len(summary) > 0 {
		for i, j := 0, len(summary); i < j; i++ {
			Echo(summary[i])
		}
	} else {
		Echo(color.YellowString("   There are no remote servers."))
	}
	Echo(line)

	l := len(c.AllList())
	max := (l - l%c.PageSize) / c.PageSize
	if l%c.PageSize > 0 {
		max++
	}
	if max == 0 {
		max = 1
	}

	Echo(strings.Repeat(" ", 7) + color.YellowString("Page: %d/%d  Total: %d", c.Page, max, l))
}

func (c *Config) exist(s string) bool {
	info, err := os.Stat(s)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func (c *Config) from(s string) error {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return err
	}
	if auth, err := c.auth(c.PrivateKey, c.Password); err != nil {
		return err
	} else {
		c.Auth = auth
	}
	for i, j := 0, len(c.Servers); i < j; i++ {
		if err = c.init(c.Servers[i]); err != nil {
			return err
		}
	}
	if c.PageSize <= 5 {
		c.PageSize = 5
	}
	c.sort()
	return nil
}

func (c *Config) init(s *Server) error {
	if s.Host == "" {
		return fmt.Errorf("the server host can not be empty")
	}
	if s.PrivateKey == "" && s.Password == "" {
		s.Auth = c.Auth
	} else {
		if auth, err := c.auth(s.PrivateKey, s.Password); err != nil {
			return err
		} else {
			s.Auth = auth
		}
	}
	if s.Port == 0 {
		s.Port = 22
	}
	if s.User == "" {
		s.User = os.Getenv("USER")
	}
	if s.Group == "" {
		s.Group = "default"
	}
	s.Addr = net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	return nil
}

func (c *Config) sort() {
	if len(c.Servers) == 0 {
		return
	}
	switch c.SortBy {
	case "name":
		sort.Slice(c.Servers, func(i, j int) bool {
			if c.Servers[i].Group == c.Servers[j].Group {
				return c.Servers[i].Name < c.Servers[j].Name
			}
			return c.Servers[i].Group < c.Servers[j].Group
		})
	case "host":
		sort.Slice(c.Servers, func(i, j int) bool {
			if c.Servers[i].Group == c.Servers[j].Group {
				return c.Servers[i].Host < c.Servers[j].Host
			}
			return c.Servers[i].Group < c.Servers[j].Group
		})
	case "disable", "":
		sort.Slice(c.Servers, func(i, j int) bool {
			return c.Servers[i].Group < c.Servers[j].Group
		})
	}
}

func (c *Config) auth(key, pass string) (ssh.AuthMethod, error) {
	if key == "" {
		return ssh.Password(pass), nil
	}
	if strings.HasPrefix(key, "~") {
		key = filepath.Join(os.Getenv("HOME"), key[1:])
	}
	data, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
