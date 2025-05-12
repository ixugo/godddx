package tmpl

type User struct {
	ID       string
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

type Sprite struct {
	ID       int
	Index    int    // 图片顺序
	Path     string // 图片存储路径
	MediaID  string // 关联的媒资 ID
	StartPTS int    // 开始 pts(秒)
	EndPTS   int    // 结束 pts(秒)
	Etag     string
	Ext      SpriteExt
}

type SpriteExt struct {
	Tiles []SpriteTile
}

type SpriteTile struct {
	Index int
	PTS   int
}
