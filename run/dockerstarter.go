package run

import (
	"context"
	"os/exec"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"

	"golang.org/x/time/rate"
)

// MaybeStartDocker tries to run Docker on the system, if not started
func MaybeStartDocker() {
	var cmd = exec.CommandContext(context.Background(), "docker", "ps")
	var err = cmd.Run()

	if err == nil {
		return
	}

	TryStartDocker()
}

// TryStartDocker tries to run Docker on the system
func TryStartDocker() {
	var wlmMsg = waitlivemsg.NewMessage("docker is starting up")
	var wlm = waitlivemsg.WaitLiveMsg{}
	wlm.SetStream(uilive.New())
	go wlm.Wait()
	defer wlmMsg.End()
	defer wlm.Stop()
	wlm.AddMessage(wlmMsg)
	var err = tryStartDocker()

	if err != nil {
		wlmMsg.SetSymbolEnd(waitlivemsg.RedCrossSymbol())
		wlmMsg.SetText("docker could not be autostarted (ignoring)")
		verbose.Debug("Ignoring error while trying to start docker: " + err.Error())
		return
	}

	var ctx, cancel = context.WithTimeout(context.Background(), time.Minute/2)
	defer cancel()

	var rlimit = rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		var cmd = exec.CommandContext(ctx, "docker", "ps")
		err = cmd.Run()

		if err == nil {
			cancel()
			wlmMsg.SetText("docker is now running")
			return
		}

		if err := rlimit.Wait(ctx); err != nil {
			wlmMsg.SetSymbolEnd(waitlivemsg.RedCrossSymbol())
			wlmMsg.SetText("docker might not be running")
			return
		}
	}
}
