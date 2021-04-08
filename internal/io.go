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
	"os"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

var State *terminal.State
var mu sync.Mutex

func init() {
	mu.Lock()
	defer mu.Unlock()
	in := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(in)
	if err != nil {
		panic(err)
	}
	State = state
	_ = terminal.Restore(in, state)
}

func Echo(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Printf("%s\r\n", format)
	} else {
		fmt.Printf("%s\r\n", fmt.Sprintf(format, args...))
	}
}

func EchoAndExit(format string, args ...interface{}) {
	Echo(format, args...)
	Exit(0)
}

func Error(format string, args ...interface{}) {
	prefix := color.New(color.FgHiRed).Sprint("ERROR")
	if len(args) == 0 {
		fmt.Printf("%s %s\r\n", prefix, color.RedString(format))
	} else {
		fmt.Printf("%s %s\r\n", prefix, color.RedString(format, args...))
	}
}

func ErrorAndExit(format string, args ...interface{}) {
	Error(format, args)
	Exit(1)
}

func ClearScreen() {
	fmt.Print("\033c")
}

func ShowTitle() {
	Echo(color.MagentaString("\r\n   J2 - A Micro Remote Server Management Client - %s\r\n", Version))
}

func Exit(n int) {
	Reset()
	os.Exit(n)
}

func Reset() {
	mu.Lock()
	if State != nil {
		_ = terminal.Restore(int(os.Stdin.Fd()), State)
	}
	mu.Unlock()
	_ = DefaultConsoleParserWrapper.TearDown()
}
