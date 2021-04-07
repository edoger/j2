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
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/c-bata/go-prompt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func Completer(doc prompt.Document) []prompt.Suggest {
	// 控制命令
	text := doc.TextBeforeCursor()
	if strings.HasPrefix(text, "-") {
		return []prompt.Suggest{
			{Text: "-n", Description: "下一页"},
			{Text: "-p", Description: "上一页"},
		}
	}
	if text == "" {
		return nil
	}
	var suggests []prompt.Suggest
	list := Cfg.PageList()
	for i, j := 0, len(list); i < j; i++ {
		suggests = append(suggests, prompt.Suggest{
			Text:        list[i].Name,
			Description: list[i].Desc,
		})
	}
	return prompt.FilterFuzzy(suggests, doc.GetWordBeforeCursor(), true)
}

func Executor(input string) {
	if input == "" {
		Cfg.ShowSummary()
		return
	}

	switch input {
	case "-n":
		Cfg.NextPage()
		Cfg.ShowSummary()
	case "-p":
		Cfg.PrevPage()
		Cfg.ShowSummary()
	default:
		list := Cfg.PageList()
		var server *Server
		for i, j := 0, len(list); i < j; i++ {
			if list[i].Name == input {
				server = list[i]
				break
			}
		}
		if server == nil {
			n, err := strconv.Atoi(input)
			if err == nil {
				n--
				if n >= 0 && n < len(list) {
					server = list[n]
				}
			}
		}
		if server == nil {
			Error("Unknown input: %s", input)
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
