package main

import (
	"fmt"
	"time"

	"github.com/henvic/uilive"
)

func main() {
	writer := uilive.New()

	for _, f := range []string{"Foo.zip", "Bar.iso"} {
		for i := 0; i <= 50; i++ {
			fmt.Fprintf(writer, "Downloading %s.. (%d/%d) GB\n", f, i, 50)
			writer.Flush()
			time.Sleep(time.Millisecond * 70)
		}

		fmt.Fprintf(writer.Bypass(), "Downloaded %s\n", f)
	}

	fmt.Fprintln(writer, "Finished: Downloaded 100GB")
}
