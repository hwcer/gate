package gate

import "github.com/hwcer/registry"

const (
	protocolTypeHTTP int8 = 1
	protocolTypeTCP  int8 = 2
	protocolTypeALL  int8 = 3
)

type protocol int8

func (p protocol) Has(t int8) bool {
	v := int8(p)
	return v|t == v
}

const Name = "gate"

type Gate struct {
	Name     string   `json:"name"`     //service name
	Prefix   string   `json:"prefix"`   //所有服务强制加前缀
	Address  string   `json:"address"`  //连接地址
	Protocol protocol `json:"protocol"` //1-短链接，2-长连接，3-长短链接全开
}

type Metadata struct {
	UID  string `json:"uid"`  //角色ID
	GUID string `json:"guid"` //账号ID
}

var Options = &struct {
	Gate     *Gate     `json:"gate"`
	Metadata *Metadata `json:"metadata"`
}{
	Gate:     &Gate{Name: "gate", Prefix: "handle", Address: "0.0.0.0:80", Protocol: 3},
	Metadata: &Metadata{UID: "uid", GUID: "guid"},
}

//HandleServicePrefixWithPath = "/" + HandleServicePrefix

// HandleServiceMethod 外部接口路径封装
func HandleServiceMethod(name string) string {
	return registry.Join(Options.Gate.Prefix, name)
}
