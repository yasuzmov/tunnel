package eventbus

import "sync"

const subscriberBuf = 32

// Bus 内存事件总线。订阅通道有界，满时丢弃该订阅者的事件以免阻塞转发路径。
type Bus struct {
	mu   sync.RWMutex
	next int
	subs map[int]chan Event
}

// New 创建事件总线。
func New() *Bus {
	return &Bus{subs: make(map[int]chan Event)}
}

// Subscribe 注册订阅者，调用方必须在结束时 Unsubscribe，否则会泄漏。
func (b *Bus) Subscribe() (int, <-chan Event) {
	ch := make(chan Event, subscriberBuf)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.next++
	id := b.next
	b.subs[id] = ch
	return id, ch
}

// Unsubscribe 关闭并移除订阅通道。
func (b *Bus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.subs[id]; ok {
		delete(b.subs, id)
		close(ch)
	}
}

// Publish 向所有订阅者投递事件。单个订阅者阻塞时跳过，不影响其他页面。
func (b *Bus) Publish(ev Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
