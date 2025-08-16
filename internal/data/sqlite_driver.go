package data
import (
	"fmt"
	"regexp"
	"strings"
	_ "github.com/glebarez/go-sqlite"
)
var (
	tokenRegex = regexp.MustCompile(`[\p{Han}]+|[A-Za-z]+|\d+|[^\s\p{Han}A-Za-z\d]+`)
)
func goTokenizeForFts(s interface{}) (string, error) {
	if s == nil {
		return "", nil
	}
	text := toString(s)
	if len(text) == 0 {
		return "", nil
	}
	var matches []string
	if tokenRegex != nil {
		matches = tokenRegex.FindAllString(text, -1)
	}
	if len(matches) == 0 {
		matches = strings.Fields(text)
	}
	tokens := make([]string, 0, len(matches)*2)
	for _, tk := range matches {
		if len(tk) == 0 {
			continue
		}
		r := []rune(tk)
		if isHan(r[0]) {
			if len(r) > 1 {
				for i := 0; i < len(r)-1; i++ {
					tokens = append(tokens, string(r[i:i+2]))
				}
			} else {
				tokens = append(tokens, tk)
			}
			continue
		}
		if isAsciiLetter(r[0]) {
			tokens = append(tokens, strings.ToLower(tk))
			continue
		}
		tokens = append(tokens, tk)
	}
	return strings.Join(tokens, " "), nil
}
func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
func isHan(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) || (r >= 0x20000 && r <= 0x2A6DF)
}
func isAsciiLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
