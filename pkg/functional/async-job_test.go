package functional

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AsyncJob(t *testing.T) {
	job := func(i int, s string) (string, error) {
		time.Sleep(time.Second)
		return fmt.Sprintf("%q--%v", s, i), nil
	}
	async := MustAsyncJob(job)
	control := async.Run(1, "QQQ")
	stat, _ := control.Wait(context.Background())
	assert.Equal(t, AsyncJobCompleted, stat.Status, "succeeded-job")
	control = async.Run("QQQ", "1")
	stat, _ = control.Wait(context.Background())
	assert.Equal(t, AsyncJobStartFailed, stat.Status, "failed-to-start-job")

	catchResult := make(chan bool, 1)
	defer close(catchResult)
	control = async.WhenCompleted(func(s string, e error) {
		select {
		case catchResult <- true:
		default:
		}
	}).Run(1, "QQQ")
	stat, _ = control.Wait(context.Background())
	assert.Equal(t, AsyncJobCompleted, stat.Status, "succeeded-job")
	assert.Equal(t, true, <-catchResult, "succeeded-job")
	control = async.WhenCompleted(func(s string, e error) {
		select {
		case catchResult <- true:
		default:
		}
	}).WhenStateChanged(func(state AsyncJobState) {
		select {
		case catchResult <- false:
		default:
		}
	}).Run("QQQ", 1)
	stat, _ = control.Wait(context.Background())
	assert.Equal(t, AsyncJobStartFailed, stat.Status, "failed-job")
	assert.Equal(t, false, <-catchResult, "failed-job")
}
