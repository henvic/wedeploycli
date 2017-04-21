package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/verbose"
)

// Stop stops the WeDeploy infrastructure
func Stop() error {
	if err := checkDockerAvailable(); err != nil {
		return err
	}

	var dm = &DockerMachine{
		Context: context.Background(),
	}
	return dm.Stop()
}

// Stop stops the machine
func (dm *DockerMachine) Stop() error {
	if dm.livew == nil {
		dm.livew = uilive.New()
	}

	dm.waitLiveMsg.ResetDuration()
	dm.waitLiveMsg.SetMessage("WeDeploy is stopping")
	dm.waitLiveMsg.SetStream(dm.livew)

	go dm.waitLiveMsg.Wait()

	if err := dm.LoadDockerInfo(); err != nil {
		return err
	}

	if dm.Container == "" {
		verbose.Debug("No infrastructure container detected.")
	}

	_ = unlinkProjects()

	if err := cleanupEnvironment(); err != nil {
		return err
	}

	dm.waitLiveMsg.StopWithMessage("WeDeploy is stopped.")

	return nil
}

func (dm *DockerMachine) beginStopListener() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go dm.stopEvent(sigs)
}

func (dm *DockerMachine) stopEvent(sigs chan os.Signal) {
	<-sigs
	verbose.Debug("WeDeploy stop event called. Waiting started signal.")
	<-dm.started
	verbose.Debug("Started end signal received.")

	dm.waitLiveMsg.Stop()

	dm.waitLiveMsg.ResetDuration()
	dm.waitLiveMsg.SetMessage("WeDeploy is stopping")
	dm.waitLiveMsg.SetStream(dm.livew)
	go dm.waitLiveMsg.Wait()

	killLoop(sigs)
	dm.terminateMutex.Lock()
	dm.terminate = true
	dm.terminateMutex.Unlock()

	_ = unlinkProjects()

	if err := cleanupEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	dm.waitLiveMsg.StopWithMessage("WeDeploy is stopped")
	dm.endRun()
}

func (dm *DockerMachine) endRun() {
	dm.end <- true
	dm.terminateMutex.RLock()
	dm.terminate = true
	dm.terminateMutex.RUnlock()
}

func (dm *DockerMachine) terminating() bool {
	return true
}

func killLoop(sigs chan os.Signal) {
	var killListenerStarted sync.WaitGroup
	killListenerStarted.Add(1)

	go func() {
		killListenerStarted.Done()
		<-sigs
		println("Cleaning up running infrastructure. Please wait.")
		<-sigs
		println("To kill this window (not recommended), try again in 60 seconds.")
		var gracefulExitLoopTimeout = time.Now().Add(1 * time.Minute)
	killLoop:
		<-sigs

		if time.Now().After(gracefulExitLoopTimeout) {
			println("\n\"we deploy\" killed awkwardly. Use \"we deploy --stop-local-infra\" to kill ghosts.")
			os.Exit(1)
		}

		goto killLoop
	}()

	killListenerStarted.Wait()
}
