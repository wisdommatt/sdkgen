package log

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

func Println(colour color.Attribute, caption string, v ...interface{}) {
	c := color.New(colour, color.Bold)
	now := time.Now()
	timezone, _ := now.Local().Zone()
	a := []interface{}{}
	a = append(a, fmt.Sprintf(
		"%d %s %d %d:%d:%d %s %s",
		now.Day(),
		now.Month().String(),
		now.Year(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		timezone,
		c.Sprint(caption),
	))
	a = append(a, v...)
	fmt.Println(a...)
}

func Fatalln(colour color.Attribute, caption string, v ...interface{}) {
	Println(colour, caption, v...)
	os.Exit(1)
}
