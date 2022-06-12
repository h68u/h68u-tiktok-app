/*
	项目定时任务，目前采用原生实现，并未使用 cron 包
*/
package cron

import (
	"sync"
	"time"

	srv "tikapp/service"
)

func Init() {
	

	// 定时更新 redis
	go func() {
		var m sync.Mutex
		for {
			func() {
				defer m.Unlock()
				time.Sleep(time.Minute * 5)
				m.Lock()
				srv.RegularUpdate()

			}()
		}
	}()

}
