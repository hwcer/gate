package gate

import (
	"github.com/hwcer/cosgo/options"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/gate/players"
	"net/http"
)

func WSVerify(w http.ResponseWriter, r *http.Request) (uid string, err error) {
	//logger.Trace("Sec-Websocket-Extensions:%v", r.Header.Get("Sec-Websocket-Extensions"))
	//logger.Trace("Sec-Websocket-Key:%v", r.Header.Get("Sec-Websocket-Key"))
	//logger.Trace("Sec-Websocket-Protocol:%v", r.Header.Get("Sec-Websocket-Protocol"))
	//logger.Trace("Sec-Websocket-Branch:%v", r.Header.Get("Sec-Websocket-Branch"))
	if !options.Gate.WSVerify {
		return "", nil
	}
	token := r.Header.Get("Sec-Websocket-Protocol")
	if token == "" {
		return "", values.Error("token empty")
	}
	sess := session.New()
	if err = sess.Verify(token); err != nil {
		return "", values.Parse(err)
	}
	uid, _ = sess.Get(options.ServiceMetadataUID).(string)
	if uid == "" {
		return "", values.Error("请登录")
	}
	return
}
func WSAccept(s *cosnet.Socket, uid string) {
	if !options.Options.Gate.WSVerify {
		return
	}
	_, _ = players.Players.Binding(s, uid, nil)
	v := values.Values{}
	v[options.ServiceMetadataUID] = uid
	s.Set(v)
	return
}
