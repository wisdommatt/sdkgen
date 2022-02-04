package log

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

func Println(colour color.Attribute, caption string, v ...interface{}) {
	c := color.New(colour, color.Bold)
	now := time.Now()
	timezone, _ := now.Local().Zone()
	a := []interface{}{}
	a = append(a, fmt.Sprintf(
		"%d %s %d %s: %s",
		now.Day(),
		now.Month().String(),
		now.Year(),
		timezone,
		c.Sprint("ERROR"),
	))
	a = append(a, v...)
	fmt.Println(a...)
}
