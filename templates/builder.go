package templates

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// New 新建builder
func New(values map[string]string) *Builder {
	return &Builder{Values: values}
}

// Builder shell 方式的模板替换,$VAR或者${VAR},支持filter${VAR|file|base64|default}
type Builder struct {
	Path   string            // 文件所在目录,file加载时使用
	Values map[string]string // 可选替换参数
}

func (b *Builder) Render(s string) string {
	builder := strings.Builder{}
	builder.Grow(2 * len(s))

	last := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			builder.WriteString(s[last:j])
			// find name
			name, w := getName(s[j+1:])
			// process name filter
			if v, ok := b.convert(name); ok {
				builder.WriteString(v)
			} else {
				builder.WriteString(s[j : j+1+w])
			}

			j += w
			last = j + 1
		}
	}

	builder.WriteString(s[last:])
	return builder.String()
}

// 通过名字映射
func (b *Builder) convert(name string) (string, bool) {
	if name == "$" {
		return "$", true
	}

	tokens := strings.Split(name, "|")
	for i, v := range tokens {
		tokens[i] = strings.TrimSpace(v)
	}

	if len(tokens) == 0 {
		return "", false
	}

	key := tokens[0]
	val := ""
	// - mean ignore variable ${-|file xxx | base64}
	if key != "-" {
		val = b.mapping(key)
		if val == "" {
			return "", false
		}
	}

	// 执行filter
	var err error
	for i := 1; i < len(tokens); i++ {
		val, err = b.filter(val, tokens[i])
		if err != nil {
			return "", false
		}
	}

	return val, true
}

func (b *Builder) mapping(key string) string {
	if v, ok := b.Values[key]; ok {
		return v
	}

	// env
	return os.Getenv(key)
}

func (b *Builder) filter(val string, processer string) (string, error) {
	fn := func(c rune) bool {
		return c == ' '
	}

	tokens := strings.FieldsFunc(processer, fn)
	if len(tokens) == 0 {
		return val, nil
	}

	key := tokens[0]
	switch key {
	case "default":
		if val == "" && len(tokens) >= 2 {
			return tokens[1], nil
		}
	case "file":
		file := val
		if val == "" && len(tokens) == 2 {
			file = tokens[1]
		}

		file = path.Join(b.Path, file)

		if file == "" {
			return "", fmt.Errorf("unknown file:%+v", processer)
		}

		data, err := ioutil.ReadFile(file)
		if err != nil {
			return val, err
		}

		return string(data), nil
	case "base64":
		return base64.StdEncoding.EncodeToString([]byte(val)), nil
	case "date":
		if val != "" {
			break
		}

		// format
		// https://programming.guide/go/format-parse-string-time-date-example.html
		return time.Now().Format("2006-01-02T15:04:05"), nil
	}

	return val, nil
}

func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// 返回名字和长度
func getName(s string) (string, int) {
	if len(s) == 0 {
		return "", 0
	}

	// $$
	if s[0] == '$' {
		return "$", 1
	}

	if s[0] == '{' {
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				return s[1:i], i + 1
			}
		}

		// not find
		return "", 1
	}

	for i := 0; i < len(s); i++ {
		if !isAlphaNum(s[i]) {
			return s[:i], i
		}
	}

	// all
	return s, len(s)
}
