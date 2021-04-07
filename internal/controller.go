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
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/c-bata/go-prompt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var CommandSuggests = []prompt.Suggest{
	{Text: "-n", Description: "Displays the next page of the server list."},
	{Text: "-p", Description: "Displays the previous page of the server list."},
	{Text: "-g", Description: "Set the group for the server list."},
	{Text: "-h", Description: "Display the usage guide of J2."},
}

func Completer(doc prompt.Document) []prompt.Suggest {
	text := doc.TextBeforeCursor()
	if text == "" {
		return nil
	}
	word := doc.GetWordBeforeCursor()
	if strings.HasPrefix(text, "-") {
		if text == "-" {
			return CommandSuggests
		}
		if strings.TrimSpace(text[1:]) == "" {
			return nil
		}
		if strings.HasPrefix(text, "-g ") {
			list := Cfg.Servers
			if len(list) == 0 {
				return nil
			}
			counts := make(map[string]int)
			for i, j := 0, len(list); i < j; i++ {
				counts[list[i].Group]++
			}
			groups := make([]string, 0, len(counts))
			for g := range counts {
				groups = append(groups, g)
			}
			if len(groups) > 1 {
				sort.Strings(groups)
			}
			suggests := make([]prompt.Suggest, 0, len(groups))
			for i, j := 0, len(groups); i < j; i++ {
				suggests = append(suggests, prompt.Suggest{
					Text:        groups[i],
					Description: fmt.Sprintf("Group %s contains %d server(s).", groups[i], counts[groups[i]]),
				})
			}
			return prompt.FilterFuzzy(suggests, word, true)
		}
		if strings.HasSuffix(text, " ") {
			return nil
		}
		return prompt.FilterHasPrefix(CommandSuggests, text, true)
	}
	if len(word) == 0 {
		return nil
	}
	list := Cfg.AllList()
	if len(list) == 0 {
		return nil
	}
	suggests := make([]prompt.Suggest, 0, len(list))
	for i, j := 0, len(list); i < j; i++ {
		suggests = append(suggests, prompt.Suggest{
			Text:        list[i].Name,
			Description: list[i].Desc,
		})
	}
	if prefixed := prompt.FilterHasPrefix(suggests, word, true); len(prefixed) == 0 {
		num, _ := strconv.Atoi(word)
		if num > 0 && num <= len(suggests) {
			return nil
		}
	}
	return prompt.FilterFuzzy(suggests, word, true)
}

func Executor(input string) {
	if input == "" {
		Cfg.ShowSummary()
		return
	}

	text := strings.TrimSpace(input)
	switch {
	case text == "-n":
		Cfg.NextPage()
		Cfg.ShowSummary()
	case text == "-p":
		Cfg.PrevPage()
		Cfg.ShowSummary()
	case strings.HasPrefix(text, "-g"):
		group := strings.TrimSpace(text[2:])
		Cfg.Group = group
		Cfg.Page = 1
		Cfg.ShowSummary()
	case text == "-h":
		ShowUsageGuide()
	default:
		var server *Server
		all := Cfg.AllList()
		for i, j := 0, len(all); i < j; i++ {
			if all[i].Name == input {
				if server != nil {
					Error("There is a remote server with the same name: %s.", input)
					return
				}
				server = all[i]
			}
		}
		if server == nil {
			list := Cfg.PageList()
			n, err := strconv.Atoi(input)
			if err == nil {
				n--
				if n >= 0 && n < len(list) {
					server = list[n]
				}
			}
		}
		if server == nil {
			Error("Instruction %q is invalid. Please use -h to view the usage guide.", input)
			return
		}
		err := Connect(server)
		Cfg.ShowSummary()
		if err != nil {
			Error("Handle server %s error: %s", server.Name, err)
		}
	}
}

func Connect(s *Server) error {
	client, err := ssh.Dial("tcp", s.Addr, &ssh.ClientConfig{
		User:            s.User,
		Auth:            []ssh.AuthMethod{s.Auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 3,
	})
	if err != nil {
		return err
	}
	defer doClose(client)

	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer doClose(sess)

	in := int(os.Stdin.Fd())
	width, height, err := terminal.GetSize(in)
	if err != nil {
		return err
	}
	state, err := terminal.MakeRaw(in)
	if err != nil {
		return err
	}
	defer func() { _ = terminal.Restore(in, state) }()

	stdin, _ := sess.StdinPipe()

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}
	err = sess.RequestPty(term, height, width, modes)
	if err != nil {
		return err
	}
	if err = sess.Shell(); err != nil {
		return err
	}

	exit := make(chan struct{})

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go loop(wg, exit, in, stdin)

	err = sess.Wait()
	close(exit)
	wg.Wait()

	if err != nil {
		switch err.(type) {
		case *ssh.ExitMissingError:
			return nil
		case *ssh.ExitError:
			return nil
		}
		return err
	}
	return nil
}

func loop(wg *sync.WaitGroup, exit chan struct{}, r int, w io.WriteCloser) {
	defer wg.Done()

	buf := make([]byte, 1024)

	for {
		select {
		case <-exit:
			return
		default:
			if n, err := syscall.Read(r, buf); err != nil {
				Error("Read input error: %s", err)
			} else {
				token := translate(buf[:n])
				if len(token) > 0 {
					_, err = w.Write(token)
					if err != nil {
						if err == io.EOF {
							return
						}
						Error("Write error: %s", err)
					}
				}
			}
		}
	}
}

func translate(token []byte) []byte {
	return token
}

func doClose(c io.Closer) {
	_ = c.Close()
}
