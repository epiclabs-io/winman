package winman

import (
	"fmt"
	"os"
)

func Log(format string, parms ...interface{}) {
	f, _ := os.OpenFile("/tmp/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintf(f, format+"\n", parms...)
}
