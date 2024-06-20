package parallel

import (
	"sync"
	"sync/atomic"

	"github.com/H-BF/corlib/pkg/lazy"
)

// ExecAbstract uses abstract algorithm to execute some task in parallel manner
/* Notes
Функция ExecAbstract - это абстрактная реализация алгоритма конкуретного выполения задач.
Считается, если задан maxConcurrency > 0, то возможно выполние задач с привлечением дополнительных (maxConcurrency) go-рутин, это
значит, что если  maxConcurrency == 0, то не будет никакого распараллеливания, все задачи будут обработаны синхронно 1-ой рутиной
вызвавшей ExecAbstract.
Если maxConcurrency == 1, то к основной рутине может быть вызавана "в помощь" еще одна дополнительная, то есть схема 1+1,
если maxConcurrency== 2, то схема 1+2, и так далее
*/
func ExecAbstract(taskCount int, maxConcurrency int32, abstractTask func(int) error) error {
	if taskCount == 0 {
		return nil
	}
	var (
		taskIndex int32
		joinPoint sync.WaitGroup
		execute   func()
	)
	errCatcher := lazy.MakeOnceInitValue()
	concurrencyLimit := int64(maxConcurrency)

	jobProduce := func(doAsk4Parallel bool) (ret func(), mayAskForParallelExecutor bool) {
		n := int(atomic.AddInt32(&taskIndex, 1) - 1)
		if n < taskCount {
			if doAsk4Parallel {
				mayAskForParallelExecutor = atomic.AddInt64(&concurrencyLimit, -1) >= 0
			}
			ret = func() {
				if err := abstractTask(n); err != nil {
					errCatcher.Assign(err)
				}
			}
		}
		return
	}

	execute = func() {
		for job, askParallelExecutor := jobProduce(true); job != nil; job, _ = jobProduce(false) {
			if e, _ := errCatcher.Get().(error); e != nil {
				break
			}
			if askParallelExecutor {
				askParallelExecutor = false
				joinPoint.Add(1)
				go func() {
					defer joinPoint.Done()
					execute()
				}()
			}
			job()
		}
	}

	execute()
	joinPoint.Wait()
	err, _ := errCatcher.Get().(error)
	return err
}

// TaskMapper ...
type TaskMapper func() func() error

// MapExecAbstract делает тоже самое, что и ExecAbstract
func MapExecAbstract(mapper TaskMapper, maxConcurrency int32) error {
	var (
		joinPoint        sync.WaitGroup
		execute          func()
		concurrencyLimit = int64(maxConcurrency)
		errCatcher       = lazy.MakeOnceInitValue()
	)
	execute = func() {
		job, canFork := mapper(), atomic.AddInt64(&concurrencyLimit, -1) >= 0
		for ; job != nil; job, canFork = mapper(), false {
			if e, _ := errCatcher.Get().(error); e != nil {
				break
			}
			if canFork {
				joinPoint.Add(1)
				go func() {
					defer joinPoint.Done()
					execute()
				}()
			}
			if e := job(); e != nil {
				errCatcher.Assign(e)
			}
		}
	}
	execute()
	joinPoint.Wait()
	err, _ := errCatcher.Get().(error)
	return err
}
