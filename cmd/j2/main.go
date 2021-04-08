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

package main

import (
	"github.com/c-bata/go-prompt"

	"github.com/edoger/j2/internal"
)

func main() {
	internal.CheckAndPrintVersion()
	internal.CheckAndPrintUsageGuide()

	if err := internal.Cfg.Init(); err != nil {
		internal.ErrorAndExit("Init config failed: %s", err)
	}

	internal.Cfg.ShowSummary()
	defer internal.Reset()

	p := prompt.New(internal.Executor, internal.Completer, internal.Options()...)
	p.Run()
}
