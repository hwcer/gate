package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"sync"
)

var Players = players{Map: sync.Map{}}

func NewPlayer(uuid string, socket *cosnet.Socket) *Player {
	return &Player{uuid: uuid, socket: socket}
}

type Player struct {
	values.Values //用户登录信息,推荐存入一个struct
	uuid          string
	mutex         sync.Mutex
	socket        *cosnet.Socket
}

type players struct {
	sync.Map
}

// replace 顶号
func (this *Player) replace(socket *cosnet.Socket) {
	var old *cosnet.Socket
	old, this.socket = this.socket, socket
	if !old.Status.Disabled() {
		old.Emit(cosnet.EventTypeReplaced)
		old.Close()
	}
	return
}

func (this *Player) UUID() string {
	return this.uuid
}

func (this *Player) Socket() *cosnet.Socket {
	return this.socket
}

func (this *players) Get(uuid string) *Player {
	v, ok := this.Load(uuid)
	if !ok {
		return nil
	}
	p, _ := v.(*Player)
	return p
}

func (this *players) Range(fn func(*Player) bool) {
	this.Map.Range(func(k, v interface{}) bool {
		if p, ok := v.(*Player); ok {
			return fn(p)
		}
		return true
	})
}

func (this *players) Delete(uuid string) bool {
	this.Map.Delete(uuid)
	return true
}

// Binding 身份认证绑定socket
func (this *players) Binding(uuid string, socket *cosnet.Socket) (r *Player, err error) {
	r = NewPlayer(uuid, socket)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if i, loaded := this.Map.LoadOrStore(uuid, r); loaded {
		r, _ = i.(*Player)
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.replace(socket)
		socket.Emit(cosnet.EventTypeReconnected)
	}
	socket.Set(r)
	socket.Emit(cosnet.EventTypeVerified)
	return
}
