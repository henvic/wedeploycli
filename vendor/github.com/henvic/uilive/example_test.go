package uilive_test

import (
	"fmt"
	"time"

	"github.com/henvic/uilive"
)

func Example() {
	writer := uilive.New()

	for i := 0; i <= 100; i++ {
		fmt.Fprintf(writer, "Downloading.. (%d/%d) GB\n", i, 100)
		writer.Flush()
		time.Sleep(time.Millisecond * 5)
	}

	fmt.Fprintln(writer.Bypass(), "Finished: Downloaded 100GB")
}
