package bellcompleter

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type BellCompleter struct {
	Completer readline.AutoCompleter
	TabCount  uint16
	Previous  string
}

func (b *BellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	runes, level := b.Completer.Do(line, pos)
	if b.Previous != string(line) {
		b.TabCount = 0
	}
	b.TabCount++
	b.Previous = string(line)

	if len(runes) == 0 {
		paths := strings.Split(os.Getenv("PATH"), ":")
		type entry struct {
			full   string
			suffix []rune
		}
		var entries []entry

		for _, path := range paths {
			if files, err := os.ReadDir(path); err == nil {
				for _, file := range files {
					if file.Type().IsRegular() && strings.HasPrefix(file.Name(), string(line)) {
						suffix, _ := strings.CutPrefix(file.Name(), string(line))
						entries = append(entries, entry{file.Name(), []rune(suffix + " ")})
					}
				}
			}
		}

		if len(entries) > 0 {
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].full < entries[j].full
			})
			for _, e := range entries {
				runes = append(runes, e.suffix)
			}
			level = len(line)
		}
	}

	if len(runes) > 1 {
		// Strip trailing spaces before computing LCP (suffixes have " " appended)
		trimmed := make([][]rune, len(runes))
		for i, r := range runes {
			trimmed[i] = []rune(strings.TrimSuffix(string(r), " "))
		}
		lcp := findLcp(trimmed)
		if len(lcp) > 0 {
			return [][]rune{[]rune(lcp)}, level
		}
		if b.TabCount == 1 {
			fmt.Print("\x07")
			return nil, 0
		}
		// TabCount == 2: show all matches
		names := make([]string, len(runes))
		for i, r := range runes {
			names[i] = string(line) + strings.TrimSuffix(string(r), " ")
		}
		fmt.Print("\n" + strings.Join(names, "  ") + "\n$ " + string(line))
		return nil, 0
	}

	if len(runes) == 0 {
		fmt.Print("\x07")
		return nil, 0
	}

	return runes, level
}

func findLcp(runes [][]rune) string {
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return string(runes[0])
	}

	minLen := len(runes[0])
	for _, r := range runes {
		if len(r) < minLen {
			minLen = len(r)
		}
	}

	var lcp []rune
	for i := 0; i < minLen; i++ {
		ch := runes[0][i]
		for _, r := range runes[1:] {
			if r[i] != ch {
				return string(lcp)
			}
		}
		lcp = append(lcp, ch)
	}
	return string(lcp)
}
