package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

const defaultWidth = 40

func main() {
	width := flag.Int("W", defaultWidth, "max line width before wrapping")
	think := flag.Bool("think", false, "use thought bubble instead of speech bubble")
	colorName := flag.String("color", "", "color name: red, green, yellow, blue, magenta, cyan, white")
	bold := flag.Bool("bold", false, "make the output bold")
	flag.Parse()

	text, err := readMessage(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "cowsay:", err)
		os.Exit(1)
	}

	out := render(text, *width, *think)
	if colored := colorize(out, *colorName, *bold); colored != "" {
		fmt.Println(colored)
		return
	}
	fmt.Println(out)
}

func readMessage(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if stat.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("no message provided (pass as args or pipe via stdin)")
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	msg := strings.TrimRight(string(data), "\n")
	if msg == "" {
		return "", fmt.Errorf("empty message on stdin")
	}
	return msg, nil
}

func render(text string, width int, think bool) string {
	lines := wrap(text, width)
	maxLen := 0
	for _, l := range lines {
		if n := visualLen(l); n > maxLen {
			maxLen = n
		}
	}

	var b strings.Builder
	b.WriteString(" " + strings.Repeat("_", maxLen+2) + "\n")
	b.WriteString(bubble(lines, maxLen, think))
	b.WriteString(" " + strings.Repeat("-", maxLen+2) + "\n")
	b.WriteString(cow(think))
	return b.String()
}

func bubble(lines []string, maxLen int, think bool) string {
	if think {
		var b strings.Builder
		for _, l := range lines {
			b.WriteString(fmt.Sprintf("( %s%s )\n", l, strings.Repeat(" ", maxLen-visualLen(l))))
		}
		return b.String()
	}

	if len(lines) == 1 {
		return fmt.Sprintf("< %s >\n", lines[0])
	}

	var b strings.Builder
	for i, l := range lines {
		pad := strings.Repeat(" ", maxLen-visualLen(l))
		var left, right byte
		switch i {
		case 0:
			left, right = '/', '\\'
		case len(lines) - 1:
			left, right = '\\', '/'
		default:
			left, right = '|', '|'
		}
		b.WriteString(fmt.Sprintf("%c %s%s %c\n", left, l, pad, right))
	}
	return b.String()
}

func cow(think bool) string {
	link := `\`
	if think {
		link = "o"
	}
	return fmt.Sprintf(`        %s   ^__^
         %s  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||`, link, link)
}

func wrap(text string, width int) []string {
	if width <= 0 {
		width = defaultWidth
	}
	var out []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			out = append(out, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := ""
		for _, w := range words {
			switch {
			case line == "" && len(w) > width:
				for len(w) > width {
					out = append(out, w[:width])
					w = w[width:]
				}
				line = w
			case line == "":
				line = w
			case len(line)+1+len(w) <= width:
				line += " " + w
			default:
				out = append(out, line)
				line = w
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func visualLen(s string) int {
	return len([]rune(s))
}

func colorize(s, name string, bold bool) string {
	attrs := []color.Attribute{}
	switch strings.ToLower(name) {
	case "":
		// no color
	case "red":
		attrs = append(attrs, color.FgRed)
	case "green":
		attrs = append(attrs, color.FgGreen)
	case "yellow":
		attrs = append(attrs, color.FgYellow)
	case "blue":
		attrs = append(attrs, color.FgBlue)
	case "magenta":
		attrs = append(attrs, color.FgMagenta)
	case "cyan":
		attrs = append(attrs, color.FgCyan)
	case "white":
		attrs = append(attrs, color.FgWhite)
	default:
		fmt.Fprintf(os.Stderr, "cowsay: unknown color %q, ignoring\n", name)
	}
	if bold {
		attrs = append(attrs, color.Bold)
	}
	if len(attrs) == 0 {
		return ""
	}
	c := color.New(attrs...)
	c.EnableColor()
	return c.Sprint(s)
}
