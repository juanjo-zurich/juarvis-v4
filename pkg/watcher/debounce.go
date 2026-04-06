package watcher

import (
	"sync"
	"time"
)

type Debouncer struct {
	window time.Duration
	mu     sync.Mutex
	buffer map[string]string
	timer  *time.Timer
	ch     chan map[string]string
}

func NewDebouncer(windowMs int) *Debouncer {
	d := &Debouncer{
		window: time.Duration(windowMs) * time.Millisecond,
		buffer: make(map[string]string),
		ch:     make(chan map[string]string),
	}
	d.resetTimer()
	return d
}

func (d *Debouncer) Add(path string, eventType string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.buffer[path] = eventType
	d.resetTimer()
}

func (d *Debouncer) Events() <-chan map[string]string {
	return d.ch
}

func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	close(d.ch)
}

func (d *Debouncer) resetTimer() {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.window, d.flush)
}

func (d *Debouncer) flush() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.buffer) == 0 {
		return
	}
	batch := make(map[string]string)
	for k, v := range d.buffer {
		batch[k] = v
	}
	d.buffer = make(map[string]string)
	d.ch <- batch
}
