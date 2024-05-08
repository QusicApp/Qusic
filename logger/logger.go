package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"fyne.io/fyne/v2/data/binding"
)

type logger struct {
	Data    *strings.Builder
	out     io.Writer
	Binding binding.String
}

func (l logger) Write(data []byte) (i int, err error) {
	l.Data.Write(data)

	l.Binding.Set(l.Data.String())

	return fmt.Fprint(l.out, *(*string)(unsafe.Pointer(&data)))
}

var Log = logger{Data: new(strings.Builder), out: os.Stdout, Binding: binding.NewString()}
var Errors = logger{Data: new(strings.Builder), out: os.Stderr, Binding: binding.NewString()}

func Print(a ...any) (i int, err error) {
	return fmt.Fprint(Log, a...)
}

func Println(a ...any) (i int, err error) {
	return fmt.Fprintln(Log, a...)
}

func Printf(format string, a ...any) (i int, err error) {
	return fmt.Fprintf(Log, format, a...)
}

func Info(a ...any) {
	fmt.Fprint(Log, "[INFO] ")
	fmt.Fprintln(Log, a...)
}

func Infof(format string, a ...any) {
	fmt.Fprint(Log, "[INFO] ")
	fmt.Fprintf(Log, format+"\n", a...)
}

func Inf(a ...any) {
	fmt.Fprint(Log, "[INFO] ")
	fmt.Fprint(Log, a...)
}

func Inff(format string, a ...any) {
	fmt.Fprint(Log, "[INFO] ")
	fmt.Fprintf(Log, format, a...)
}

func Error(a ...any) {
	fmt.Fprint(Errors, "[ERROR] ")
	fmt.Fprintln(Errors, a...)
}

func Errorf(format string, a ...any) {
	fmt.Fprint(Errors, "[ERROR] ")
	fmt.Fprintf(Errors, format+"\n", a...)
}

func Err(a ...any) {
	fmt.Fprint(Errors, "[ERROR] ")
	fmt.Fprint(Errors, a...)
}

func Errf(format string, a ...any) {
	fmt.Fprint(Errors, "[ERROR] ")
	fmt.Fprintf(Errors, format, a...)
}
