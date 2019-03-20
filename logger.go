package xormigrate

import (
	"io/ioutil"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "xormigrate: ", 0)

func (x *Xormigrate) SetLogger(l *log.Logger) {
	if l != nil {
		logger = l
	} else {
		logger = log.New(ioutil.Discard, "", 0)
	}
}
