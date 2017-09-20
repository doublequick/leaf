package timer

import (
	"runtime"
	"time"

	"github.com/doublequick/leaf/conf"
	"github.com/doublequick/leaf/log"
)

// TimerTask 任务定时器
type TimerTask struct {
	t  *time.Timer
	cb func()
}

// Stop 停止任务定时器
func (t *TimerTask) Stop() {
	t.t.Stop()
	t.cb = nil
}

// Cb 调用任务定时器任务
func (t *TimerTask) Cb() {
	defer func() {
		t.cb = nil
		if r := recover(); r != nil { // 如果panic则记录出错信息
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	if t.cb != nil {
		t.cb()
	}
}

// Cron 计划任务
type Cron struct {
	t *TimerTask
}

// Stop 停止计划任务
func (c *Cron) Stop() {
	if c.t != nil {
		c.t.Stop()
	}
}

// Dispatcher 调度程序, one dispatcher per goroutine (goroutine not safe)
type Dispatcher struct {
	ChanTimerTask chan *TimerTask
}

// NewDispatcher 新建调度程序, l == 0 同步, l > 0 缓存channel
func NewDispatcher(l int) *Dispatcher {
	disp := new(Dispatcher)
	disp.ChanTimerTask = make(chan *TimerTask, l)
	return disp
}

// AfterFunc 定时`d`后执行`cb`
func (disp *Dispatcher) AfterFunc(d time.Duration, cb func()) *TimerTask {
	timerTask := new(TimerTask)
	timerTask.cb = cb
	timerTask.t = time.AfterFunc(d, func() {
		disp.ChanTimerTask <- timerTask
	})
	return timerTask
}

// CronFunc 计划调度
func (disp *Dispatcher) CronFunc(cronExpr *CronExpr, _cb func()) *Cron {
	c := new(Cron)

	now := time.Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return c
	}

	// callback
	var cb func()
	cb = func() {
		defer _cb()

		now := time.Now()
		nextTime := cronExpr.Next(now)
		if nextTime.IsZero() {
			return
		}
		c.t = disp.AfterFunc(nextTime.Sub(now), cb)
	}

	c.t = disp.AfterFunc(nextTime.Sub(now), cb)
	return c
}
