package file

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jeckbjy/kubegen/templates"
	"gopkg.in/yaml.v2"
)

// Doc 一个文档
type Doc struct {
	ID   string        // 扩展ID,用于合并?
	Root yaml.MapSlice // 所有根节点
}

// File 一个yaml文件,可能包含多个文档
type File struct {
	Docs []*Doc
}

// Len 文档个数
func (f *File) Len() int {
	return len(f.Docs)
}

// GetDoc 返回文档根节点
func (f *File) GetDoc(i int) yaml.MapSlice {
	return f.Docs[i].Root
}

// ProcessAll 加载所有文件,并替换文本
func (f *File) ProcessAll(files []string, expand bool, values map[string]string) ([]byte, error) {
	builder := templates.New(values)
	for _, filename := range files {
		layer := File{}
		if err := layer.Load(filename); err != nil {
			return nil, err
		}

		if expand {
			layer.Expand()
		}

		// convert
		builder.Path = filepath.Dir(filename)
		layer.Build(builder)

		if err := f.Concat(&layer); err != nil {
			return nil, err
		}
	}

	return f.Marshal()
}

// Load 加载一个yaml文件,并保持顺序
func (f *File) Load(filename string) error {
	f.Docs = f.Docs[:0]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	chunks := bytes.Split(data, []byte("\n---\n"))
	for _, chunk := range chunks {
		doc := yaml.MapSlice{}
		err := yaml.Unmarshal(chunk, &doc)
		if err != nil {
			return err
		}

		f.Docs = append(f.Docs, &Doc{Root: doc})
	}

	return nil
}

// Expand 展开以点分隔的key节点
func (f *File) Expand() {
	for _, doc := range f.Docs {
		if normalize(doc.Root) {
			doc.Root = concat(yaml.MapSlice{}, doc.Root)
		}
	}
}

// Marshal 序列化yaml文件
func (f *File) Marshal() ([]byte, error) {
	builder := bytes.Buffer{}
	for idx, doc := range f.Docs {
		out, err := yaml.Marshal(doc.Root)
		if err != nil {
			return nil, err
		}

		if idx > 0 {
			builder.WriteString("\n---\n")
		}

		builder.Write(out)
	}

	return builder.Bytes(), nil
}

// Concat 合并两个yaml文件,要求必须是都只含有一个document
func (f *File) Concat(other *File) error {
	if other.Len() == 0 {
		return nil
	}

	if f.Len() == 0 {
		f.Docs = other.Docs
		return nil
	}

	if f.Len() > 1 || other.Len() > 1 {
		return errors.New("concat file do not support multiple document in one file")
	}

	f.Docs[0].Root = concat(f.GetDoc(0), other.GetDoc(0))

	return nil
}

// Build 替换文本
func (f *File) Build(builder *templates.Builder) {
	for _, doc := range f.Docs {
		render(builder, doc.Root)
	}
}

// render 递归替换文本
func render(builder *templates.Builder, doc interface{}) {
	switch doc.(type) {
	case yaml.MapSlice:
		slice := doc.(yaml.MapSlice)
		for idx := range slice {
			item := &slice[idx]
			if str, ok := item.Value.(string); ok {
				item.Value = builder.Render(str)
			} else {
				render(builder, item.Value)
			}
		}
	case []interface{}:
		slice := doc.([]interface{})
		for idx := range slice {
			item := slice[idx]
			if str, ok := item.(string); ok {
				slice[idx] = builder.Render(str)
			} else {
				render(builder, item)
			}
		}
	case map[interface{}]interface{}:
		mm := doc.(map[interface{}]interface{})
		for key, val := range mm {
			if str, ok := val.(string); ok {
				mm[key] = builder.Render(str)
			} else {
				render(builder, val)
			}
		}
	default: // int 类型等
		//fmt.Printf("unknown type:%+v\n", doc)
	}
}

// 格式化yaml文件,将以点分隔的key展开
func normalize(doc yaml.MapSlice) bool {
	processed := false
	// 单纯展开即可
	for idx := 0; idx < len(doc); idx++ {
		item := &doc[idx]
		if slice, ok := item.Value.(yaml.MapSlice); ok {
			if normalize(slice) {
				processed = true
			}
		}

		if key, ok := item.Key.(string); ok && strings.Contains(key, ".") {
			processed = true
			tokens := strings.Split(key, ".")
			value := item.Value
			last := item
			for i := 0; i < len(tokens); i++ {
				if i == len(tokens)-1 {
					last.Key = tokens[i]
					last.Value = value
				} else {
					s := yaml.MapSlice{}
					s = append(s, yaml.MapItem{})
					last.Key = tokens[i]
					last.Value = s
					last = &s[0]
				}
			}
		}
	}

	return processed
}

// concat 合并两个文档
func concat(f1 yaml.MapSlice, f2 yaml.MapSlice) yaml.MapSlice {
	for _, e2 := range f2 {
		key := e2.Key.(string)

		found := false
		for i1, e1 := range f1 {
			if e1.Key.(string) == key {
				found = true
				// 要求e1和e2的value都得是MapSlice类型
				s1, ok1 := e1.Value.(yaml.MapSlice)
				s2, ok2 := e2.Value.(yaml.MapSlice)
				if !ok1 || !ok2 {
					panic("conncat fail:not mapslice")
				}
				// use reference
				f1[i1].Value = concat(s1, s2)
				//e1.Value = concat(s1, s2)
				break
			}
		}

		if !found {
			f1 = append(f1, yaml.MapItem{Key: key, Value: e2.Value})
		}
	}

	return f1
}
