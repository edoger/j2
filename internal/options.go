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

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

var KeyHandlers = map[prompt.Key]prompt.KeyBindFunc{
	prompt.ControlC: DoExit,
	prompt.Enter:    DoEnter,
	prompt.ControlM: DoEnter,
	prompt.ControlJ: DoEnter,
	prompt.PageUp:   DoPrevPage,
	prompt.PageDown: DoNextPage,
	prompt.Home:     DoFirstPage,
	prompt.End:      DoLastPage,
}

func DoExit(*prompt.Buffer) {
	EchoAndExit(color.HiGreenString("Bye~"))
}

func DoEnter(*prompt.Buffer) {
	Cfg.ShowSummary()
}

func DoPrevPage(*prompt.Buffer) {
	Cfg.PrevPage()
	Cfg.ShowSummary()
}

func DoNextPage(*prompt.Buffer) {
	Cfg.NextPage()
	Cfg.ShowSummary()
}

func DoFirstPage(*prompt.Buffer) {
	Cfg.Page = 1
	Cfg.ShowSummary()
}

func DoLastPage(*prompt.Buffer) {
	Cfg.Page = 1
	Cfg.PrevPage()
	Cfg.ShowSummary()
}

func DoMakeLivePrefix() (string, bool) {
	if Cfg.Group != "" {
		return fmt.Sprintf("j2 [%s] >> ", Cfg.Group), true
	}
	return "", false
}

func Options() []prompt.Option {
	binds := make([]prompt.KeyBind, 0, len(KeyHandlers))
	for key, fn := range KeyHandlers {
		binds = append(binds, prompt.KeyBind{Key: key, Fn: fn})
	}

	return []prompt.Option{
		prompt.OptionTitle("J2 - A Micro Remote Server Management Client"),
		prompt.OptionPrefix("j2 >> "),
		prompt.OptionAddKeyBind(binds...),
		prompt.OptionPrefixTextColor(prompt.Blue),
		prompt.OptionPrefixBackgroundColor(prompt.DefaultColor),
		prompt.OptionSuggestionTextColor(prompt.Brown),
		prompt.OptionSuggestionBGColor(prompt.DefaultColor),
		prompt.OptionSelectedSuggestionTextColor(prompt.Red),
		prompt.OptionSelectedSuggestionBGColor(prompt.Yellow),
		prompt.OptionDescriptionTextColor(prompt.Cyan),
		prompt.OptionDescriptionBGColor(prompt.DefaultColor),
		prompt.OptionSelectedDescriptionTextColor(prompt.Fuchsia),
		prompt.OptionSelectedDescriptionBGColor(prompt.Yellow),
		prompt.OptionCompletionOnDown(),
		prompt.OptionLivePrefix(DoMakeLivePrefix),
	}
}
