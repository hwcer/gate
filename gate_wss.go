package gate

import (
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosnet"
	"github.com/hwcer/gate/players"
	"github.com/hwcer/wower/options"
	"net/http"
)

func WSVerify(w http.ResponseWriter, r *http.Request) (uuid string, err error) {
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
	uuid = sess.UUID()
	return
}
func WSAccept(s *cosnet.Socket, uuid string) {
	if !options.Options.Gate.WSVerify {
		return
	}
	_, _ = players.Players.Binding(s, uuid, nil)
	return
}
