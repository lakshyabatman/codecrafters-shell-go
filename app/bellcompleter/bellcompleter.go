package bellcompleter

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type BellCompleter struct {
	Completer readline.AutoCompleter
}

func (b *BellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	runes, level := b.Completer.Do(line, pos)
	fmt.Println(level)
	if len(runes) == 0 {
		paths := strings.Split(os.Getenv("PATH"), ":")
		candidates := [][]rune{}
		for _, path := range paths {
			if files, err := os.ReadDir(path); err == nil {
				for _, file := range files {
					if file.Type().IsRegular() && strings.HasPrefix(file.Name(), string(line)) {
						if len(file.Name()) > 0 {
							candidates = append(candidates, []rune(file.Name()))

						}

					}
				}
			}
			if len(candidates) > 0 {
				// fmt.Println(candidates)
				return candidates, len(line)
			}
		}

	}

	if len(runes) == 0 {
		fmt.Print("\x07")
	}

	return runes, level
}
