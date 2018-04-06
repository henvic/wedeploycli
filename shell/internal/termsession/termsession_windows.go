// +build windows

package termsession

import "time"

func (t *TermSession) start() {}

func (t *TermSession) watchResize() {
	t.Resize()

	// sleep for 250ms like Kubernete's kubectl does, instead of handling Windows signals.
	// k8s.io/kubernetes/pkg/kubectl/util/term/resizeevents_windows.go#L58-L59
	// commit: fc8bfe2d8929e11a898c4557f9323c482b5e8842
	ticker := time.NewTicker(250 * time.Millisecond)

	for {
		select {
		case <-t.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			t.Resize()
		}
	}
}

func (t *TermSession) restore() {}
