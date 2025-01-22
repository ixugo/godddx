package tmpl

type User struct {
	Age      int64  //  年龄
	Name     string // 昵称
	Password []byte
	UserLogs UserLogs
}

type UserLogs struct {
	ID     int
	Action string // 行为
	IP     string // 操作 ip
}
