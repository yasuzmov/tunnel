// Package logwriter 异步批量写入请求日志。
// 转发路径只往有界 channel 投递，写满时丢弃并计数，绝不阻塞访客请求。
package logwriter

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// Entry 一条待持久化的请求日志及其可选 body。
type Entry struct {
	Log  *model.RequestLog
	Body *model.RequestLogBody
}

// Writer 后台刷盘器。
type Writer struct {
	ch       chan *Entry
	repo     repository.LogRepository
	bus      *eventbus.Bus
	batch    int
	interval time.Duration
	dropped  atomic.Int64
}

// New 创建写入器。
func New(repo repository.LogRepository, bus *eventbus.Bus, channelSize, batch int, interval time.Duration) *Writer {
	if channelSize <= 0 {
		channelSize = 4096
	}
	if batch <= 0 {
		batch = 200
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &Writer{
		ch:       make(chan *Entry, channelSize),
		repo:     repo,
		bus:      bus,
		batch:    batch,
		interval: interval,
	}
}

// Enqueue 投递日志。通道满时丢弃当前条。
func (w *Writer) Enqueue(e *Entry) {
	if e == nil || e.Log == nil {
		return
	}
	select {
	case w.ch <- e:
	default:
		n := w.dropped.Add(1)
		if n%100 == 1 {
			log.Printf("请求日志通道已满，累计丢弃 %d 条", n)
		}
	}
}

// Dropped 返回累计丢弃数。
func (w *Writer) Dropped() int64 {
	return w.dropped.Load()
}

// Run 直到 ctx 取消。退出前会尽量把剩余条目刷完。
func (w *Writer) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	buf := make([]*Entry, 0, w.batch)
	flush := func(flushCtx context.Context) {
		if len(buf) == 0 {
			return
		}
		w.flush(flushCtx, buf)
		buf = buf[:0]
	}
	for {
		select {
		case <-ctx.Done():
			flush(context.Background())
			w.drain()
			return
		case <-ticker.C:
			flush(ctx)
		case e := <-w.ch:
			buf = append(buf, e)
			if len(buf) >= w.batch {
				flush(ctx)
			}
		}
	}
}

func (w *Writer) drain() {
	for {
		select {
		case e := <-w.ch:
			w.flush(context.Background(), []*Entry{e})
		default:
			return
		}
	}
}

func (w *Writer) flush(ctx context.Context, batch []*Entry) {
	logs := make([]*model.RequestLog, 0, len(batch))
	bodies := make([]*model.RequestLogBody, 0, len(batch))
	for _, e := range batch {
		logs = append(logs, e.Log)
		if e.Body != nil {
			bodies = append(bodies, e.Body)
		} else {
			bodies = append(bodies, &model.RequestLogBody{})
		}
	}
	if err := w.repo.CreateBatch(ctx, logs, bodies); err != nil {
		log.Printf("批量写入请求日志失败: %v", err)
		return
	}
	if w.bus == nil {
		return
	}
	for _, item := range logs {
		w.bus.Publish(eventbus.Event{Type: eventbus.EventLogCreated, Payload: item})
	}
}
