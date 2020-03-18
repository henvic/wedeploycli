package listinstances

import (
	"fmt"
	"strings"

	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/formatter"
	"github.com/henvic/wedeploycli/services"
	"github.com/henvic/wedeploycli/verbose"
)

// Printf list
func (l *List) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(l.w, format, a...)
}

func (l *List) printInstances() {
	l.selectors = []string{}

	l.watchMutex.RLock()
	var instances = l.Instances
	l.watchMutex.RUnlock()

	if len(instances) == 0 {
		l.Printf("No instance found.\n")
		return
	}

	header := "Instance\tState"

	if l.SelectNumber {
		header = "#\t" + header
	}

	if formatter.Human {
		header = strings.Replace(header, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, header))

	for _, instance := range instances {
		l.printInstance(instance)
	}
}

func (l *List) printInstance(instance services.Instance) {
	if l.SelectNumber {
		l.selectors = append(l.selectors, instance.ContainerID)

		l.Printf("%d\t", len(l.selectors))
	}

	printed := instance.ContainerID

	if len(printed) > 12 && !verbose.Enabled {
		printed = printed[:11]
	}

	l.Printf("%s\t%s\n", printed, instance.State)
}
