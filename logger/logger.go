package logger

import (
	"fmt"
	"os"
)

func Info(a ...any) {
	fmt.Print("[INFO] ")
	fmt.Println(a...)
}

func Infof(format string, a ...any) {
	fmt.Print("[INFO] ")
	fmt.Printf(format+"\n", a...)
}

func Inf(a ...any) {
	fmt.Print("[INFO] ")
	fmt.Print(a...)
}

func Inff(format string, a ...any) {
	fmt.Print("[INFO] ")
	fmt.Printf(format, a...)
}

func Error(a ...any) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprintln(os.Stderr, a...)
}

func Errorf(format string, a ...any) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

func Err(a ...any) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprint(os.Stderr, a...)
}

func Errf(format string, a ...any) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
	fmt.Fprintf(os.Stderr, format, a...)
}
