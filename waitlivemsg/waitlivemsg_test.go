package waitlivemsg

import (
	"testing"
	"time"

	"github.com/henvic/uilive"
)

func TestWaitLiveMsg(t *testing.T) {
	var us = uilive.New()
	var wlm = WaitLiveMsg{}
	wlm.SetStream(us)
	var one = NewMessage("foo")
	var two = NewMessage("bar")

	wlm.AddMessage(one)
	wlm.AddMessage(two)
	go wlm.Wait()
	time.Sleep(2 * time.Second)
	one.PlayText("xyz")
	time.Sleep(3 * time.Second)
	wlm.SetMessage(two)
	time.Sleep(3 * time.Second)
	var three = NewMessage("iziiz three")
	wlm.AddMessage(three)
	time.Sleep(2 * time.Second)
	three.StopText("iziiz stopped")
	time.Sleep(2 * time.Second)
	three.PlayText("iziiz continued")
	time.Sleep(2 * time.Second)
	wlm.RemoveMessage(three)
	time.Sleep(2 * time.Second)
	wlm.Stop()
}

func TestWaitLiveMsgStopNoTick(t *testing.T) {
	var wlm = WaitLiveMsg{}

	end := make(chan struct{}, 1)
	timer := time.NewTimer(time.Second)

	go func() {
		wlm.Stop()
		end <- struct{}{}
	}()

	select {
	case <-timer.C:
		t.Errorf("WaitLiveMsg didn't stop before tick")
	case <-end:
	}
}
