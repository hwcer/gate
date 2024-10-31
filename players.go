package gate

import (
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"sync"
)

var Players = players{Map: sync.Map{}}

func NewPlayer(uuid string) *Player {
	return &Player{uuid: uuid}
}

type Player struct {
	values.Values        //用户登录信息,推荐存入一个struct
	uuid          string //GUID
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
	if old != nil && old.Id() == socket.Id() {
		return
	}
	if old != nil && !old.Status.Disabled() {
		old.Emit(cosnet.EventTypeReplaced)
		old.Close()
	}
	//合并数据
	//for k, v := range socket.Values() {
	//	this.Values.Set(k, v)
	//}
	//socket.Set(this.Values)
	return
}

func (this *Player) UUID() string {
	return this.uuid
}

func (this *Player) Socket() *cosnet.Socket {
	return this.socket
}

func (this *Player) Get(key string) interface{} {
	return this.Values.Get(key)
}
func (this *Player) Set(key string, value any) any {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	vs := values.Values{}
	vs.Merge(this.Values, true)
	vs.Set(key, value)
	this.Values = vs
	return value
}

// Merge 批量设置Cookie信息
func (this *Player) Merge(data map[string]any, replace bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	vs := values.Values{}
	vs.Merge(this.Values, false)
	vs.Merge(data, replace)
	this.Values = vs
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

type loginCallback func(player *Player, loaded bool) error

func (this *players) login(uuid string, data values.Values, callback loginCallback) error {
	p := NewPlayer(uuid)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Values = values.Values{}
	i, loaded := this.Map.LoadOrStore(uuid, p)
	if loaded {
		p, _ = i.(*Player)
		p.mutex.Lock()
		defer p.mutex.Unlock()
		p.Values.Merge(data, true)
	} else {
		p.Values.Merge(data, true)
	}
	return callback(p, loaded)
}

func (this *players) Login(uuid string, data values.Values) (r *Player, err error) {
	err = this.login(uuid, data, func(player *Player, _ bool) error {
		r = player
		return nil
	})
	return
}

// Binding 身份认证绑定socket
func (this *players) Binding(uuid string, socket *cosnet.Socket, data values.Values) (r *Player, err error) {
	err = this.login(uuid, data, func(player *Player, loaded bool) error {
		player.socket = socket
		if loaded {
			r.replace(socket)
			socket.Emit(cosnet.EventTypeReconnected)
		} else {
			socket.Emit(cosnet.EventTypeConnected)
		}
		r = player
		return nil
	})
	return
}
