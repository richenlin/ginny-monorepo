package retry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/syncmap"
)

type ExeFunc = func(ctx context.Context, param interface{}) (interface{}, error)

// ConcurrencyRetryCallFunc
func ConcurrencyRetryCallFunc(ctx context.Context, fs map[string]ExeFunc,
	params map[string]interface{}, retry ...int) (map[string]interface{}, error) {
	var (
		wg      = new(sync.WaitGroup)
		errChan = make(chan error, len(fs))
		resMap  = syncmap.Map{}
	)
	for k, f := range fs {
		wg.Add(1)
		go func(c context.Context, k string, f ExeFunc, p interface{}) {
			var err error
			defer func(err *error) {
				if *err != nil {
					errChan <- *err
				}
				wg.Done()
			}(&err)
			var res interface{}
			res, err = RetryCallFunc(c, f, p, retry...)
			resMap.Store(k, res)
		}(ctx, k, f, params[k])
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}
	result := map[string]interface{}{}
	resMap.Range(func(k, v interface{}) bool {
		result[k.(string)] = v
		return true
	})
	return result, nil
}

// RetryCallFunc
// retry  retryTimes, retryPeriod, timeout
func RetryCallFunc(ctx context.Context, fn ExeFunc,
	params interface{}, retry ...int) (interface{}, error) {
	var (
		err         error
		result      interface{}
		retryTimes  int = 1
		retryPeriod     = 100 * time.Millisecond
		timeout         = time.Millisecond * 600000 // 单次最大时长默认10分钟
	)
	l := len(retry)
	if l == 1 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
	} else if l == 2 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
		if retry[1] > 0 {
			retryPeriod = time.Duration(retry[1]) * time.Millisecond
		}
	} else if l > 2 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
		if retry[1] > 0 {
			retryPeriod = time.Duration(retry[1]) * time.Millisecond
		}
		if retry[2] > 0 {
			timeout = time.Duration(retry[2]) * time.Millisecond
		}
	}

	num := 0
	o := time.After(timeout)
	t := time.NewTicker(retryPeriod)
	defer t.Stop()
	for range t.C {
		select {
		case <-o:
			return nil, fmt.Errorf("call function context deadline exceeded, params: [%+v]", params)
		default:
		}
		num++
		result, err = fn(ctx, params)
		if err != nil {
			if num < retryTimes {
				time.Sleep(retryPeriod)
				continue
			}
			return nil, fmt.Errorf("err:[%v]", err.Error())
		}
		break
	}
	return result, nil
}
