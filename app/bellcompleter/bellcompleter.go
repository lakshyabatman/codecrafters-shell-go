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

func resolveDir(prefix string) string {
	if strings.ContainsAny(prefix, "/") {
		return "./" + prefix[:strings.LastIndex(prefix, "/")]
	}
	return "."
}
func resolveFile(prefix string) string {
	if strings.ContainsAny(prefix, "/") {
		return prefix[strings.LastIndex(prefix, "/")+1:]
	}
	return prefix
}

func findFileMatches(prefix string) []string {

	dir := resolveDir(prefix)
	fileToComplete := resolveFile(prefix)
	entries, err := os.ReadDir(dir)
	// fmt.Print(dir + " ")
	// fmt.Println(entries)
	if err != nil {
		return nil
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		// fmt.Println(name + " " + fileToComplete)
		if strings.HasPrefix(name, fileToComplete) {
			if entry.IsDir() {
				matches = append(matches, name+"/")
			} else {
				matches = append(matches, name+" ")
			}

		}
	}
	return matches

	// sort.Strings(matches)
	// return matches
}

func (b *BellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if b.Previous != string(line) {
		b.TabCount = 0
	}

	b.TabCount++
	b.Previous = string(line)

	if strings.ContainsAny(string(line), " \t") {
		input := string(line)
		lastSpace := strings.LastIndexAny(input, " \t")
		prefix := input[lastSpace+1:]
		if len(prefix) == 0 {
			if b.TabCount == 2 {
				matches := findFileMatches("./")
				names := make([]string, len(matches))
				for i, r := range matches {
					names[i] = string(line) + strings.TrimSuffix(string(r), " ")
				}
				fmt.Print("\n" + strings.Join(names, "  ") + "\n$ " + string(line))
				return nil, 0
			}
			return nil, 0
		}
		matches := findFileMatches(prefix)
		if len(matches) == 0 {
			fmt.Print("\a")
			return nil, 0
		}

		if len(matches) == 1 {
			rest := matches[0][len(resolveFile(prefix)):]
			return [][]rune{[]rune(rest)}, len(prefix)
		}

		fmt.Print("\a")
		return nil, 0
	}

	runes, level := b.Completer.Do(line, pos)
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
