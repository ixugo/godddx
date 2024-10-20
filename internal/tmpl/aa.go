package tmpl

import (
	"git.lnton.com/lnton/pkg/orm"
	"github.com/lib/pq"
)

type User struct {
	Name string
	Age  int64
	Age2 int16
}

// Channel 通道
type Channel struct {
	ID         string         `gorm:"primaryKey;column:id" json:"id"` // ID
	CreatedAt  orm.Time       `gorm:"type:timestamptz;notNull;default:CURRENT_TIMESTAMP;index;comment:创建时间" json:"created_at"`
	UpdatedAt  orm.Time       `gorm:"type:timestamptz;notNull;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	Enabled    bool           `gorm:"column:enabled;notNull;default:true;comment:是否启用" json:"enabled"`         // 是否启用
	Name       string         `gorm:"column:name;notNull;default:'';comment:通道名称" json:"name"`                 // 通道名称
	DeviceID   string         `gorm:"column:device_id;notNull;default:'';index;comment:设备ID" json:"device_id"` // 设备 id
	Protocol   string         `gorm:"column:protocol;notNull;default:'';comment:通道协议" json:"protocol"`         // 通道协议
	PTZType    int            `gorm:"column:ptz_type;notNull;default:0;comment:云台类型" json:"ptz_type"`          // 云台类型
	Remark     string         `gorm:"column:remark;notNull;default:'';comment:备注" json:"remark"`               // 备注描述
	Transport  string         `gorm:"column:transport;notNull;default:'TCP';comment:传输协议" json:"transport"`    // TCP/UDP
	IP         string         `gorm:"column:ip;notNull;default:'';comment:IP" json:"ip"`                       // ip 地址
	Port       int            `gorm:"column:port;notNull;default:0;comment:端口号" json:"port"`                   // 端口号
	Username   string         `gorm:"column:username;notNull;default:'';comment:用户名" json:"-"`                 // 用户名
	Password   string         `gorm:"column:password;notNull;default:'';comment:密码" json:"-"`                  // 密码
	BID        string         `gorm:"column:bid;notNull;default:'';comment:协议专属 id" json:"bid"`
	PTZ        bool           `gorm:"column:ptz;notNull;default:FALSE;comment:是否支持 ptz" json:"ptz"` // 是否支持 ptz
	Talk       bool           `gorm:"column:talk;notNull;default:FALSE;comment:是否支持对讲" json:"talk"` // 是否支持语音对讲
	PID        string         `gorm:"column:pid;notNull;index;default:'';comment:父通道 ID" json:"pid"`
	Groups     pq.StringArray `gorm:"column:groups;type:text[];default:'{}';comment:虚拟组织" json:"-"`
	Ext        ChannelExt     `gorm:"column:ext;type:jsonb;notNull;default:'{}';comment:扩展字段" json:"ext"`
	ChildCount int            `gorm:"column:child_count;notNull;default:0" json:"child_count"` // 子通道数量(不包含子孙通道)
	Status     bool           `gorm:"column:status;notNull;default:false;comment:通道状态" json:"status"`

	URL                 string `gorm:"-" json:"url"`                   //
	Playing             bool   `gorm:"-" json:"playing"`               // 是否播放中
	IsDir               bool   `gorm:"-" json:"is_dir"`                // 是否目录
	RecordPlanEnabled   bool   `gorm:"-" json:"record_plan_enabled"`   // 是否启用了录像计划
	CascadeShareEnabled bool   `gorm:"-" json:"cascade_share_enabled"` // 是否启用了级联共享
	CustomID            string `gorm:"-" json:"custom_id"`
	// Latitude  float32
	// Longitude float32
}

// Channel 通道
// type Channel struct {
// 	ID         string         `gorm:"primaryKey;column:id" json:"id"` // ID
// 	CreatedAt  orm.Time       `gorm:"type:timestamptz;notNull;default:CURRENT_TIMESTAMP;index;comment:创建时间" json:"created_at"`
// 	UpdatedAt  orm.Time       `gorm:"type:timestamptz;notNull;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
// 	Enabled    bool           `gorm:"column:enabled;notNull;default:true;comment:是否启用" json:"enabled"`         // 是否启用
// 	Name       string         `gorm:"column:name;notNull;default:'';comment:通道名称" json:"name"`                 // 通道名称
// 	DeviceID   string         `gorm:"column:device_id;notNull;default:'';index;comment:设备ID" json:"device_id"` // 设备 id
// 	Protocol   string         `gorm:"column:protocol;notNull;default:'';comment:通道协议" json:"protocol"`         // 通道协议
// 	PTZType    int            `gorm:"column:ptz_type;notNull;default:0;comment:云台类型" json:"ptz_type"`          // 云台类型
// 	Remark     string         `gorm:"column:remark;notNull;default:'';comment:备注" json:"remark"`               // 备注描述
// 	Transport  string         `gorm:"column:transport;notNull;default:'TCP';comment:传输协议" json:"transport"`    // TCP/UDP
// 	IP         string         `gorm:"column:ip;notNull;default:'';comment:IP" json:"ip"`                       // ip 地址
// 	Port       int            `gorm:"column:port;notNull;default:0;comment:端口号" json:"port"`                   // 端口号
// 	Username   string         `gorm:"column:username;notNull;default:'';comment:用户名" json:"-"`                 // 用户名
// 	Password   string         `gorm:"column:password;notNull;default:'';comment:密码" json:"-"`                  // 密码
// 	BID        string         `gorm:"column:bid;notNull;default:'';comment:协议专属 id" json:"bid"`
// 	PTZ        bool           `gorm:"column:ptz;notNull;default:FALSE;comment:是否支持 ptz" json:"ptz"` // 是否支持 ptz
// 	Talk       bool           `gorm:"column:talk;notNull;default:FALSE;comment:是否支持对讲" json:"talk"` // 是否支持语音对讲
// 	PID        string         `gorm:"column:pid;notNull;index;default:'';comment:父通道 ID" json:"pid"`
// 	Groups     pq.StringArray `gorm:"column:groups;type:text[];default:'{}';comment:虚拟组织" json:"-"`
// 	Ext        ChannelExt     `gorm:"column:ext;type:jsonb;notNull;default:'{}';comment:扩展字段" json:"ext"`
// 	ChildCount int            `gorm:"column:child_count;notNull;default:0" json:"child_count"` // 子通道数量(不包含子孙通道)
// 	Status     bool           `gorm:"column:status;notNull;default:false;comment:通道状态" json:"status"`

// 	URL                 string `gorm:"-" json:"url"`                   //
// 	Playing             bool   `gorm:"-" json:"playing"`               // 是否播放中
// 	IsDir               bool   `gorm:"-" json:"is_dir"`                // 是否目录
// 	RecordPlanEnabled   bool   `gorm:"-" json:"record_plan_enabled"`   // 是否启用了录像计划
// 	CascadeShareEnabled bool   `gorm:"-" json:"cascade_share_enabled"` // 是否启用了级联共享
// 	CustomID            string `gorm:"-" json:"custom_id"`
// 	// Latitude  float32
// 	// Longitude float32
// }

// ChannelExt 通道的扩展内容，主要集中在国标上
type ChannelExt struct {
	Parental int    `json:"parental"`  // 是否有子设备(1:有;0:没有)
	ParentID string `json:"parent_id"` // 父设备/区域/系统ID
}
