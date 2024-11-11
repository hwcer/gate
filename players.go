package gate

import (
	"github.com/hwcer/cosnet"
	"github.com/hwcer/cosweb/session"
	"sync"
)

const (
	SessionPlayerSocketName = "player.sock"
)

var Players = players{Map: sync.Map{}}

//func NewPlayer(p *session.Player) *Player {
//	sp := &Player{Player: p}
//	return sp
//}
//
//type Player struct {
//	*session.Player
//	socket *cosnet.Socket
//}

type players struct {
	sync.Map
}

// replace 顶号
func (this *players) replace(p *session.Player, socket *cosnet.Socket) {
	old := this.Socket(p)
	p.Set(SessionPlayerSocketName, socket)
	if old == nil || old.Id() == socket.Id() {
		return
	}
	if !old.Status.Disabled() {
		old.Emit(cosnet.EventTypeReplaced)
		old.Close()
	}
	return
}

func (this *players) Socket(p *session.Player) *cosnet.Socket {
	i := p.Get(SessionPlayerSocketName)
	if i == nil {
		return nil
	}
	r, _ := i.(*cosnet.Socket)
	return r
}

func (this *players) Get(uuid string) *session.Player {
	v, ok := this.Load(uuid)
	if !ok {
		return nil
	}
	p, _ := v.(*session.Player)
	return p
}

func (this *players) Range(fn func(*session.Player) bool) {
	this.Map.Range(func(k, v interface{}) bool {
		if p, ok := v.(*session.Player); ok {
			return fn(p)
		}
		return true
	})
}

func (this *players) Delete(p *session.Player) bool {
	if p == nil {
		return false
	}
	this.Map.Delete(p.UUID())
	return true
}

type loginCallback func(player *session.Player, loaded bool) error

func (this *players) Login(p *session.Player, callback loginCallback) (err error) {
	p.Lock()
	defer p.Unlock()
	r := p
	i, loaded := this.Map.LoadOrStore(p.UUID(), p)
	if loaded {
		sp, _ := i.(*session.Player)
		sp.Merge(p)
		r = sp
	}
	if callback != nil {
		err = callback(r, loaded)
	}
	return
}

// Binding 身份认证绑定socket
func (this *players) Binding(socket *cosnet.Socket, uuid string, data map[string]any) (r *session.Player, err error) {
	p := session.NewPlayer(uuid, string(socket.Id()), data)
	err = this.Login(p, func(player *session.Player, loaded bool) error {
		if loaded {
			this.replace(player, socket)
			socket.Emit(cosnet.EventTypeReconnected)
		} else {
			player.Set(SessionPlayerSocketName, socket)
			socket.Emit(cosnet.EventTypeConnected)
		}
		r = player
		return nil
	})
	return
}
