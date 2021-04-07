package internal

import (
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

var KeyHandlers = map[prompt.Key]prompt.KeyBindFunc{
	prompt.ControlC: DoExit,
	prompt.Enter:    DoEnter,
	prompt.ControlM: DoEnter,
}

func DoExit(*prompt.Buffer) {
	EchoAndExit(color.HiGreenString(" Bye~"))
}

func DoEnter(*prompt.Buffer) {
	ClearScreen()
	Cfg.ShowSummary()
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
	}
}