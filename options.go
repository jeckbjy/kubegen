package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/jeckbjy/kubegen/file"
)

func NewOptions() *Options {
	return &Options{values: make(map[string]string)}
}

// Group 文件集合
type Group struct {
	File   string   // 主文件
	Layers []string // 需要合并的文件
}

// Options 解析命令行参数
type Options struct {
	apply     bool              // 是否调用kubectl apply
	expand    bool              // 是否将点分隔的key展开
	input     string            // 输入目录
	output    string            // 输出目录
	prefix    string            // 输出前缀
	suffix    string            // 输出后缀
	namespace string            // apply需要的命名空间
	config    string            // values配置文件
	selector  string            // values选择标识
	values    map[string]string // value集合
	groups    []Group           // 需要处理的文件
}

// Parse 解析命令行
func (op *Options) Parse() {
	if op.values == nil {
		op.values = make(map[string]string)
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--apply":
			op.apply = true
		case "--expand":
			op.expand = true
		case "--prefix":
			i++
			op.prefix = getArg(i)
		case "--suffix":
			i++
			op.suffix = getArg(i)
		case "-i":
			i++
			op.input = getArg(i)
		case "-o":
			i++
			op.output = getArg(i)
		case "-n":
			i++
			op.namespace = getArg(i)
		case "-c":
			i++
			op.config = getArg(i)
		case "-s":
			i++
			op.selector = getArg(i)
		case "-v":
			i++
			op.parseValue(getArg(i))
		case "-l":
			i++
			layer, index := op.parseLayer(getArg(i))
			op.AddLayers(index, layer)
		default:
			op.AddFiles(arg)
		}
	}
}

// Process 执行参数
func (op *Options) Process() {
	if op.values == nil {
		op.values = make(map[string]string)
	}

	// load files
	if len(op.groups) == 0 {
		panic("need at least one yaml file")
	}

	// parse input value files
	op.loadValueFile()

	for _, group := range op.groups {
		f := file.File{}
		files := make([]string, 0, len(group.Layers)+1)
		files = append(files, group.File)
		files = append(files, group.Layers...)
		if op.input != "" {
			for i, f := range files {
				files[i] = path.Join(op.input, f)
			}
		}

		data, err := f.ProcessAll(files, op.expand, op.values)
		if err != nil {
			panic(err)
		}

		if op.output != "" {
			op.WriteFile(files[0], data)
		} else {
			fmt.Printf("%s", data)
		}

		op.Apply(data)
	}
}

func (op *Options) Apply(data []byte) {
	if !op.apply {
		return
	}

	// kubectl apply -n ${NAME_SPACE} -f -
	args := []string{"apply"}
	if op.namespace != "" {
		args = append(args, []string{"-n", op.namespace}...)
	}
	args = append(args, []string{"-f", "-"}...)
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if stdin, err := cmd.StdinPipe(); err == nil {
		stdin.Write(data)
	}
	cmd.Run()
}

// AddValue 添加一个value
func (op *Options) AddValue(key string, val string) {
	op.values[key] = val
}

// AddFiles 添加文件
func (op *Options) AddFiles(files ...string) {
	for _, file := range files {
		op.groups = append(op.groups, Group{File: file})
	}
}

func (op *Options) AddLayers(index int, layers ...string) error {
	if index >= len(op.groups) {
		return fmt.Errorf("bad file group:%+v", index)
	}

	op.groups[index].Layers = append(op.groups[index].Layers, layers...)
	return nil
}

func (op *Options) parseLayer(data string) (string, int) {
	index := strings.LastIndexByte(data, '=')
	if index == -1 {
		return data, 0
	}

	layer := data[:index]
	groupIdx, err := strconv.Atoi(strings.TrimSpace(data[index+1:]))
	if err != nil {
		panic(fmt.Errorf("bad layer index:%+v", data))
		return "", -1
	}

	return layer, groupIdx
}

func (op *Options) parseValue(data string) {
	tokens := strings.SplitN(data, "=", 2)
	if len(tokens) != 2 {
		return
	}

	op.values[tokens[0]] = tokens[1]
}

func (op *Options) loadValueFile() {
	filename := op.config

	if filename == "" {
		return
	}

	if op.input != "" && !strings.Contains(filename, "/") {
		filename = filepath.Join(op.input, filename)
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	doc := yaml.MapSlice{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		panic(err)
	}

	for _, item := range doc {
		if sub, ok := item.Value.(yaml.MapSlice); ok {
			// second level
			if str, ok := item.Key.(string); ok && str == op.selector {
				for _, e := range sub {
					key := fmt.Sprintf("%+v", e.Key)
					val := fmt.Sprintf("%+v", e.Value)
					op.values[key] = val
				}
			}
		} else {
			key := fmt.Sprintf("%+v", item.Key)
			val := fmt.Sprintf("%+v", item.Value)
			op.values[key] = val
		}
	}
}

func (op *Options) WriteFile(filename string, data []byte) {
	name := filepath.Base(filename)

	if op.prefix != "" {
		name = op.prefix + name
	}

	if op.suffix != "" {
		ext := filepath.Ext(filename)
		name = strings.TrimSuffix(name, ext)
		name = name + op.suffix + ext
	}

	if op.output != "" && op.output != "." {
		os.MkdirAll(op.output, os.ModePerm)
	}

	out := filepath.Join(op.output, name)
	err := ioutil.WriteFile(out, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func getArg(index int) string {
	if index >= len(os.Args) {
		panic("args index overflow")
	}

	return os.Args[index]
}
