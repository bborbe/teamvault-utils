package main

import (
	"testing"

	"bytes"

	. "github.com/bborbe/assert"
)

func TestDo(t *testing.T) {
	var err error
	writer := bytes.NewBufferString("")
	input := bytes.NewBufferString(" http://www.example.com ")
	err = do(writer, input)
	err = AssertThat(err, NilValue())
	if err != nil {
		t.Fatal(err)
	}
	err = AssertThat(writer.String(), Is("http://www.example.com\n"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoBugfix(t *testing.T) {
	var err error
	writer := bytes.NewBufferString("")
	input := bytes.NewBufferString("\n2015-04-01T15:40:09 http://www.example.com\n2015-04-01T15:40:09 http://www.example.com\n")
	err = do(writer, input)
	err = AssertThat(err, NilValue())
	if err != nil {
		t.Fatal(err)
	}
	err = AssertThat(writer.String(), Is("http://www.example.com\nhttp://www.example.com\n"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoWithParameters(t *testing.T) {
	var err error
	writer := bytes.NewBufferString("")
	input := bytes.NewBufferString(" http://www.example.com?a=b ")
	err = do(writer, input)
	err = AssertThat(err, NilValue())
	if err != nil {
		t.Fatal(err)
	}
	err = AssertThat(writer.String(), Is("http://www.example.com?a=b\n"))
	if err != nil {
		t.Fatal(err)
	}
}
