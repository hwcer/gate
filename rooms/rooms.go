package rooms

import (
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosrpc/xshare"
	"github.com/hwcer/gate/players"
	"strings"
	"sync"
)

var rooms = sync.Map{}

func Get(name string) (r *Room) {
	if i, ok := rooms.Load(name); ok {
		r = i.(*Room)
	}
	return
}
func loadOrCreate(name string) (r *Room, loaded bool) {
	i, loaded := rooms.LoadOrStore(name, &Room{})
	r = i.(*Room)
	return
}

func Join(name string, p *session.Data) {
	for _, k := range strings.Split(name, ",") {
		if room, _ := loadOrCreate(k); room != nil {
			room.Join(p)
		}
	}
}

func Leave(name string, p *session.Data) {
	for _, k := range strings.Split(name, ",") {
		if room := Get(k); room != nil {
			room.Leave(p)
		}
	}
}

func Range(name string, f func(*session.Data) bool) {
	room := Get(name)
	if room == nil {
		return
	}
	room.Range(f)
}

func Broadcast(c *xshare.Context) any {
	path := c.GetMetadata(options.ServiceMessagePath)
	name := c.GetMetadata(options.ServiceMessageRoom)
	room := Get(name)
	if room == nil {
		return false
	}

	ignore := c.GetMetadata(options.ServiceMessageIgnore)
	ignoreMap := make(map[string]struct{})
	if ignore != "" {
		arr := strings.Split(ignore, ",")
		for _, v := range arr {
			ignoreMap[v] = struct{}{}
		}
	}
	body := c.Bytes()

	for _, p := range room.ps {
		uid := p.GetString(options.ServiceMetadataUID)
		if _, ok := ignoreMap[uid]; ok {
			socket := players.Socket(p)
			if socket != nil {
				_ = socket.Send(path, body)
			}
		}
	}
	return true
}
