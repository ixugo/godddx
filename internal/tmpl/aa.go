package tmpl

import "github.com/ixugo/goweb/pkg/orm"

type User struct {
	Age      int64  //  年龄
	Name     string // 昵称
	Password []byte
}

type UserLogs struct {
	ID        int
	CreatedAt orm.Time // 创建时间
	UpdatedAt orm.Time // 更新时间
	Action    string   // 行为
	IP        string   // 操作 ip
}
