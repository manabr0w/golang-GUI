package painter

import (
	"image"
	"sync"

	"golang.org/x/exp/shiny/screen"
)

type Receiver interface {
	Update(t screen.Texture)
}

type Loop struct {
	Receiver Receiver

	next screen.Texture
	prev screen.Texture

	Mq      MessageQueue
	stopped chan struct{}
	stopReq bool
}

var size = image.Pt(800, 800)

func (l *Loop) Start(s screen.Screen) {
	l.next, _ = s.NewTexture(size)
	l.prev, _ = s.NewTexture(size)

	l.stopped = make(chan struct{})
	go func() {
		for !l.stopReq || !l.Mq.Empty() {
			op := l.Mq.Pull()
			update := op.Do(l.next)
			if update {
				l.Receiver.Update(l.next)
				l.next, l.prev = l.prev, l.next
			}
		}
		close(l.stopped)
	}()
}

func (l *Loop) Post(op Operation) {
	l.Mq.Push(op)
}

func (l *Loop) StopAndWait() {
	l.Post(OperationFunc(func(screen.Texture) {
		l.stopReq = true
	}))
	<-l.stopped
}

type MessageQueue struct {
	Ops     []Operation
	mu      sync.Mutex
	blocked chan struct{}
}

func (mq *MessageQueue) Push(op Operation) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.Ops = append(mq.Ops, op)
	if mq.blocked != nil {
		close(mq.blocked)
		mq.blocked = nil
	}
}

func (mq *MessageQueue) Pull() Operation {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	for len(mq.Ops) == 0 {
		mq.blocked = make(chan struct{})
		mq.mu.Unlock()
		<-mq.blocked
		mq.mu.Lock()
	}
	op := mq.Ops[0]
	mq.Ops[0] = nil
	mq.Ops = mq.Ops[1:]
	return op
}

func (mq *MessageQueue) Empty() bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	return len(mq.Ops) == 0
}
