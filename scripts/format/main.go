package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: format_themes <file...>")
		os.Exit(2)
	}

	for _, arg := range os.Args[1:] {
		if err := formatFile(arg); err != nil {
			fmt.Fprintf(os.Stderr, "format %s: %v\n", arg, err)
			os.Exit(1)
		}
	}
}

func formatFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data any
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	var f formatter
	if err := f.formatValue(data, 0); err != nil {
		return err
	}
	out := f.String()

	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(out), mode)
}

type formatter struct {
	b strings.Builder
}

func (f *formatter) String() string {
	out := f.b.String()
	if out == "" || out[len(out)-1] != '\n' {
		out += "\n"
	}
	return out
}

func (f *formatter) writeIndent(n int) {
	if n <= 0 {
		return
	}
	f.b.WriteString(strings.Repeat(" ", n))
}

func (f *formatter) writeLine(s string) {
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}

func (f *formatter) formatValue(v any, indent int) error {
	switch val := v.(type) {
	case map[string]any:
		f.writeIndent(indent)
		f.writeLine("{")
		if err := f.formatMapBody(val, indent); err != nil {
			return err
		}
		f.writeIndent(indent)
		f.writeLine("}")
		return nil
	case []any:
		f.writeIndent(indent)
		f.writeLine("[")
		if err := f.formatSliceBody(val, indent); err != nil {
			return err
		}
		f.writeIndent(indent)
		f.writeLine("]")
		return nil
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		f.writeIndent(indent)
		f.writeLine(string(b))
		return nil
	}
}

func (f *formatter) formatMapBody(m map[string]any, indent int) error {
	orderedKeys := orderedGroupedKeys(m)
	for i, k := range orderedKeys {
		f.writeIndent(indent + 2)
		f.b.WriteString(quote(k))
		f.b.WriteString(": ")

		switch val := m[k].(type) {
		case map[string]any:
			f.writeLine("{")
			if err := f.formatMapBody(val, indent+2); err != nil {
				return err
			}
			f.writeIndent(indent + 2)
			f.writeLine("}")
		case []any:
			f.writeLine("[")
			if err := f.formatSliceBody(val, indent+2); err != nil {
				return err
			}
			f.writeIndent(indent + 2)
			f.writeLine("]")
		default:
			b, err := json.Marshal(val)
			if err != nil {
				return err
			}
			f.writeLine(string(b))
		}

		if i < len(orderedKeys)-1 {
			f.trimTrailingNewline()
			f.b.WriteString(",\n")
		}

		if i < len(orderedKeys)-1 {
			currPrefix := keyPrefix(k)
			nextPrefix := keyPrefix(orderedKeys[i+1])
			if currPrefix != nextPrefix {
				f.writeLine("")
			}
		}
	}
	return nil
}

func (f *formatter) formatSliceBody(items []any, indent int) error {
	for i, item := range items {
		if err := f.formatValue(item, indent+2); err != nil {
			return err
		}

		if i < len(items)-1 {
			f.trimTrailingNewline()
			f.b.WriteString(",\n")
		}
	}
	return nil
}

func orderedGroupedKeys(m map[string]any) []string {
	groups := map[string][]string{}
	for k := range m {
		prefix := keyPrefix(k)
		groups[prefix] = append(groups[prefix], k)
	}

	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	var ordered []string
	for _, name := range groupNames {
		keys := groups[name]
		sort.Strings(keys)
		ordered = append(ordered, keys...)
	}
	return ordered
}

func keyPrefix(key string) string {
	if !strings.Contains(key, ".") {
		// Treat non-prefixed keys as a single root group.
		return ""
	}
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func quote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return "\"\""
	}
	return string(b)
}

func (f *formatter) trimTrailingNewline() {
	if f.b.Len() == 0 {
		return
	}
	out := f.b.String()
	if out[len(out)-1] == '\n' {
		f.b.Reset()
		f.b.WriteString(out[:len(out)-1])
	}
}
