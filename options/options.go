package options

import "github.com/hwcer/cosrpc/xshare"

const (
	ProtocolTypeHTTP int8 = 1
	ProtocolTypeTCP  int8 = 2
	ProtocolTypeALL  int8 = 3
)

const Name = "gate"

type protocol int8

func (p protocol) Has(t int8) bool {
	v := int8(p)
	return v|t == v
}

type Gate struct {
	Address   string   `json:"address"`   //连接地址
	Protocol  protocol `json:"protocol"`  //1-短链接，2-长连接，3-长短链接全开
	Broadcast int8     `json:"broadcast"` //Push message 0-关闭，1-双向通信，2-独立启动服务器,推送消息必须启用长链接
}
type Metadata struct {
	API  string `json:"api"`  //socket 推送消息时的路径(协议)
	UID  string `json:"uid"`  //角色ID
	GUID string `json:"guid"` //账号ID
}

var Options = &struct {
	Appid    string            `json:"appid"` //项目标识
	Rpcx     *xshare.Rpcx      `json:"rpcx"`
	Gate     *Gate             `json:"gate"`
	Service  map[string]string `json:"service"`
	Metadata *Metadata         `json:"metadata"`
}{
	Rpcx:     xshare.Options,
	Gate:     &Gate{Address: "0.0.0.0:80", Protocol: 3, Broadcast: 1},
	Service:  xshare.Service,
	Metadata: &Metadata{API: "api", UID: "uid", GUID: "guid"},
}
