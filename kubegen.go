package main

import (
	"fmt"
	"os"
	"strings"
)

// kubegen service.yaml --apply -o out/service.yaml -i values.yaml -s aws -v APP=commgame -v ENV=alpha
func main() {
	if len(os.Args) <= 1 || os.Args[1] == "help" {
		builder := strings.Builder{}
		builder.WriteString("usage:\n")
		builder.WriteString("  kubegen <files> [-l filename=index] [--apply] [--expand] [-i input_dir] [-o output_dir] [--prefix value] [--suffix value] [-c values_file] [-s selector] [-v key=value] [-n namespace]\n")
		fmt.Printf(builder.String())
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "%+v", r)
		}
	}()

	options := Options{}
	options.Parse()
	options.Process()
}
