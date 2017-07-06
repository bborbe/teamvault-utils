package main

import (
	"bytes"
	"io"
	"os"

	"flag"

	crawler_linkparser "github.com/bborbe/crawler/linkparser"

	"fmt"

	"runtime"

	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	writer := os.Stdout
	input := os.Stdin
	err := do(writer, input)
	if err != nil {
		glog.Exit(err)
	}
}

func do(writer io.Writer, input io.Reader) error {
	contentBuffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(contentBuffer, input); err != nil {
		return err
	}
	linkparser := crawler_linkparser.New()
	links := linkparser.ParseAbsolute(string(contentBuffer.Bytes()))
	for match := range links {
		fmt.Fprintf(writer, "%s\n", match)
	}

	return nil
}
