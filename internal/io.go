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

	"github.com/fatih/color"
)

func Echo(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Printf("%s\r\n", format)
	} else {
		fmt.Printf("%s\r\n", fmt.Sprintf(format, args...))
	}
}

func EchoAndExit(format string, args ...interface{}) {
	Echo(format, args...)
	os.Exit(0)
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
	os.Exit(1)
}

func ClearScreen() {
	fmt.Print("\033c")
}

func ShowTitle() {
	Echo(color.MagentaString("\r\n   J2 - A Micro Remote Server Management Client - %s\r\n", Version))
}
