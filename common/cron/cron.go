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
	var m sync.Mutex


	// 定时更新 redis
	go func() {
		for {
			time.Sleep(time.Second * 5)
			m.Lock()
			srv.RegularUpdate()
			m.Unlock()
		}
	}()
	

}