package rooms

import (
	"github.com/hwcer/cosgo/session"
	"sync"
)

type Room struct {
	ps     map[string]*session.Data
	locker sync.Mutex
}

func (this *Room) Has(v *session.Data) bool {
	_, ok := this.ps[v.UUID()]
	return ok
}

func (this *Room) Join(d *session.Data) {
	if this.Has(d) {
		return
	}
	this.locker.Lock()
	defer this.locker.Unlock()
	ps := map[string]*session.Data{}
	for k, v := range this.ps {
		ps[k] = v
	}
	ps[d.UUID()] = d
	this.ps = ps
}

func (this *Room) Leave(d *session.Data) {
	if !this.Has(d) {
		return
	}
	this.locker.Lock()
	defer this.locker.Unlock()
	delete(this.ps, d.UUID())
}

func (this *Room) Range(f func(*session.Data) bool) {
	for _, p := range this.ps {
		if !f(p) {
			return
		}
	}
}
