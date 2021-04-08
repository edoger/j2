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
	"strings"

	"github.com/fatih/color"
)

func ShowUsageGuide() {
	Echo("")
	var size int
	for i, j := 0, len(CommandSuggests); i < j; i++ {
		if n := len(CommandSuggests[i].Text); n > size {
			size = n
		}
	}
	format := fmt.Sprintf("%%-%ds", size)
	var texts []string
	for i, j := 0, len(CommandSuggests); i < j; i++ {
		texts = append(texts, "  "+fmt.Sprintf(format, CommandSuggests[i].Text)+"  "+CommandSuggests[i].Description)
	}
	texts = append(texts, "")
	texts = append(texts, "* Enter the number/name and press <Enter> to automatically connect to")
	texts = append(texts, "  the corresponding remote server.")
	texts = append(texts, "* Use <Control+D> to exit J2.")

	prefix := strings.Repeat(" ", 5)
	Echo(prefix + color.GreenString("J2 Usage Guide:"))
	Echo("")
	for i, j := 0, len(texts); i < j; i++ {
		if texts[i] == "" {
			Echo("")
		} else {
			Echo(prefix + color.GreenString(texts[i]))
		}
	}
	Echo("")
}

func CheckAndPrintUsageGuide() {
	for i, j := 1, len(os.Args); i < j; i++ {
		switch os.Args[i] {
		case "--help", "-help", "-h":
			ShowUsageGuide()
			prefix := strings.Repeat(" ", 5)
			Echo(prefix + color.GreenString("Command Args:"))
			Echo(prefix + color.GreenString("  -h, -help, --help"))
			Echo(prefix + color.GreenString("    Print this message and exit."))
			Echo(prefix + color.GreenString("  -v, -version, --version"))
			Echo(prefix + color.GreenString("    Print J2 version and exit."))
			Echo("")
			Exit(0)
		}
	}
}
