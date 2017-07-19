package pullimages

import (
	"fmt"
	"strings"

	"github.com/wedeploy/cli/metrics"
)

type pullMetrics struct {
	image string
}

func (p *pullMetrics) reportStart() {
	p.report("pull_image_start",
		fmt.Sprintf("Starting to pull service image %v", p.image))
}

func (p *pullMetrics) reportError() {
	p.report("pull_image_error",
		fmt.Sprintf("Error while trying to pull service image %v", p.image))
}

func (p *pullMetrics) reportSuccess() {
	p.report("pull_image_success",
		fmt.Sprintf("Service image %v successfully pulled", p.image))
}

func (p *pullMetrics) report(metricsType, text string) {
	var tags = []string{}

	if strings.HasSuffix(p.image, "wedeploy/") {
		tags = append(tags, "official_image")
	}

	metrics.Rec(metrics.Event{
		Type: metricsType,
		Text: text,
		Tags: tags,
		Extra: map[string]string{
			"image_type": p.image,
		},
	})
}
