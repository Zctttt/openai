package models

import (
	"go-chatgpt-api/common/md5"
	"math/rand"
	"time"
)

// 会员个人信息
type MtUser struct {
	Id          int        `gorm:"column:ID;type:int(11);AUTO_INCREMENT;comment:会员ID;NOT NULL" json:"ID"`
	MOBILE      string     `gorm:"column:MOBILE;type:varchar(20);comment:手机号码" json:"MOBILE"`
	USERNO      string     `gorm:"column:USER_NO;type:varchar(30);comment:会员号" json:"USER_NO"`
	AVATAR      string     `gorm:"column:AVATAR;type:varchar(255);comment:头像" json:"AVATAR"`
	NAME        string     `gorm:"column:NAME;type:varchar(30);comment:称呼" json:"NAME"`
	OPENID      string     `gorm:"column:OPEN_ID;type:varchar(50);comment:微信open_id" json:"OPEN_ID"`
	IDCARD      string     `gorm:"column:IDCARD;type:varchar(20);comment:证件号码" json:"IDCARD"`
	GRADEID     string     `gorm:"column:GRADE_ID;type:varchar(10);default:1;comment:等级ID" json:"GRADE_ID"`
	STARTTIME   *time.Time `gorm:"column:START_TIME;type:datetime;comment:会员开始时间" json:"START_TIME"`
	ENDTIME     *time.Time `gorm:"column:END_TIME;type:datetime;comment:会员结束时间" json:"END_TIME"`
	BALANCE     float64    `gorm:"column:BALANCE;type:float(10,2);default:0.00;comment:余额" json:"BALANCE"`
	POINT       int32      `gorm:"column:POINT;type:int(11);default:0;comment:积分" json:"POINT"`
	SEX         int32      `gorm:"column:SEX;type:int(11);default:0;comment:性别 0男；1女" json:"SEX"`
	BIRTHDAY    string     `gorm:"column:BIRTHDAY;type:varchar(20);comment:出生日期" json:"BIRTHDAY"`
	CARNO       string     `gorm:"column:CAR_NO;type:varchar(10);comment:车牌号" json:"CAR_NO"`
	SOURCE      string     `gorm:"column:SOURCE;type:varchar(30);comment:来源渠道" json:"SOURCE"`
	PASSWORD    string     `gorm:"column:PASSWORD;type:varchar(32);comment:密码" json:"PASSWORD"`
	SALT        string     `gorm:"column:SALT;type:varchar(4);comment:salt" json:"SALT"`
	ADDRESS     string     `gorm:"column:ADDRESS;type:varchar(100);comment:地址" json:"ADDRESS"`
	STOREID     int32      `gorm:"column:STORE_ID;type:int(11);default:0;comment:默认店铺" json:"STORE_ID"`
	CREATETIME  time.Time  `gorm:"column:CREATE_TIME;type:datetime;comment:创建时间" json:"CREATE_TIME"`
	UPDATETIME  time.Time  `gorm:"column:UPDATE_TIME;type:datetime;comment:更新时间" json:"UPDATE_TIME"`
	STATUS      string     `gorm:"column:STATUS;type:char(1);default:A;comment:状态，A：激活；N：禁用；D：删除" json:"STATUS"`
	DESCRIPTION string     `gorm:"column:DESCRIPTION;type:varchar(255);comment:备注信息" json:"DESCRIPTION"`
	OPERATOR    string     `gorm:"column:OPERATOR;type:varchar(30);comment:最后操作人" json:"OPERATOR"`
	UNIONID     string     `gorm:"column:UNION_ID;type:varchar(50);NOT NULL" json:"UNION_ID"`
}

func (m *MtUser) TableName() string {
	return "mt_user"
}

func NewUser() *MtUser {
	return &MtUser{}
}

func (m *MtUser) SetName(name string) {
	m.NAME = name
}

func (m *MtUser) SetPassword(password string) {
	m.setSalt()
	m.PASSWORD = md5.GetMD5(password + m.SALT)
}

func (m *MtUser) ValidatePassword(password string) bool {
	return md5.ValidateMD5(password+m.SALT, m.PASSWORD)
}

func (m *MtUser) setSalt() {
	rand.Seed(time.Now().UnixNano())

	// 生成4个随机字符
	randomChars := make([]byte, 4)
	for i := range randomChars {
		randomChars[i] = byte(rand.Intn(26) + 65)
	}
	m.SALT = string(randomChars)
}

func (m *MtUser) InitTime() {
	m.CREATETIME = time.Now()
	m.UPDATETIME = time.Now()
}
