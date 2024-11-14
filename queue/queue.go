// Package queue
package queue

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mnhkahn/gogogo/logger"
	"github.com/mnhkahn/gogogo/panicer"
)

type Queue struct {
	queue *list.List
	cap   int
	mu    sync.RWMutex

	gap    time.Duration
	popFn  func(e interface{})
	popLen int
}

var (
	ErrQueueFull = errors.New("queue is full")
)

func NewQueue(cap int, gap time.Duration, popLen int, f func(e interface{})) (*Queue, error) {
	if f == nil {
		return nil, fmt.Errorf("f can't be nil")
	}
	if gap <= 0 {
		gap = 20 * time.Second
	}
	if popLen <= 0 {
		popLen = 100
	}
	q := new(Queue)

	q.queue = list.New()
	q.cap = cap
	q.gap = gap
	q.popFn = f
	q.popLen = popLen

	go q.consumer()

	return q, nil
}

func (q *Queue) consumer() {
	defer panicer.Recover()
	logger.Info("start msgQueue consumer len:", q.gap.String())

	timer := time.NewTimer(q.gap)
	for {
		select {
		case <-timer.C:
			l, err := q.popFront(q.popLen, q.popFn)
			if err != nil {
				logger.Warn("pop front error:", err.Error())
			}
			logger.Debug("pop front len:", l)

			if q.Len() > q.popLen {
				timer.Reset(1 * time.Second)
			} else {
				timer.Reset(q.gap)
			}
		}
	}
}

func (q *Queue) Push(e interface{}) error {
	if q.queue.Len() >= q.cap {
		return ErrQueueFull
	}

	q.mu.Lock()
	q.queue.PushBack(e)
	q.mu.Unlock()

	return nil
}

func (q *Queue) PushFront(e interface{}) error {
	if q.queue.Len() >= q.cap {
		return ErrQueueFull
	}

	q.mu.Lock()
	q.queue.PushFront(e)
	q.mu.Unlock()

	return nil
}

func (q *Queue) popFront(l int, f func(e interface{})) (int, error) {
	if l > q.Len() {
		l = q.Len()
	}

	for i := 0; i < l; i++ {
		q.mu.Lock()
		e := q.queue.Front()
		if e != nil {
			q.queue.Remove(e)
		}
		q.mu.Unlock()

		if e != nil {
			f(e.Value)
		}
	}

	return l, nil
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.queue.Len()
}

func (q *Queue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue.Init()

	return nil
}

func (q *Queue) Debug(l int) []interface{} {
	res := make([]interface{}, 0, l)

	if l >= q.Len() || l == 0 {
		l = q.Len()
	}
	e := q.queue.Front()
	for i := 0; i < l && e != nil; i++ {
		res = append(res, e.Value)
		e = e.Next()
	}

	return res
}

func (q *Queue) PopHandler(w http.ResponseWriter, r *http.Request) {
	l, err := q.popFront(q.popLen, q.popFn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(strconv.Itoa(l)))
}

func (q *Queue) DebugHandler(w http.ResponseWriter, r *http.Request) {
	l, _ := strconv.Atoi(r.URL.Query().Get("l"))
	buf, err := json.Marshal(q.Debug(l))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}
