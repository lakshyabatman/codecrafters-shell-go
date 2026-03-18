package bellcompleter

import (
	"fmt"

	"github.com/chzyer/readline"
)

type BellCompleter struct {
	Completer readline.AutoCompleter
}

func (b *BellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	runes, level := b.Completer.Do(line, pos)
	if len(runes) == 0 {
		fmt.Print("\x07")
	}
	return runes, level
}
