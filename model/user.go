package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrevicenzi/unchained"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	Id            int64     `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:用户 id;"`
	Username      string    `form:"username" json:"username,omitempty" gorm:"column:username;type:varchar(64);size:64;index:username,unique;comment:用户名;"`
	Password      string    `form:"password" json:"password,omitempty" gorm:"column:password;type:varchar(128);size:128;comment:密码;"`
	Nickname      string    `form:"nickname" json:"nickname,omitempty" gorm:"column:nickname;type:varchar(64);size:64;comment:用户昵称;"`
	Mobile        string    `form:"mobile" json:"mobile,omitempty" gorm:"column:mobile;type:varchar(20);size:20;index:mobile;comment:手机号;"`
	Email         string    `form:"email" json:"email,omitempty" gorm:"column:email;type:varchar(64);size:64;index:email;comment:联系邮箱;"`
	Address       string    `form:"address" json:"address,omitempty" gorm:"column:address;type:varchar(255);size:255;comment:联系地址;"`
	Signature     string    `form:"signature" json:"signature,omitempty" gorm:"column:signature;type:varchar(255);size:255;comment:签名;"`
	LastLoginIp   string    `form:"last_login_ip" json:"last_login_ip,omitempty" gorm:"column:last_login_ip;type:varchar(32);size:32;comment:最后登录 ip 地址;"`
	RegisterIp    string    `form:"register_ip" json:"register_ip,omitempty" gorm:"column:register_ip;type:varchar(32);size:32;comment:注册ip;"`
	DocCount      int       `form:"doc_count" json:"doc_count,omitempty" gorm:"column:doc_count;type:int(10) unsigned;default:0;comment:上传的文档数;"`
	FollowCount   int       `form:"follow_count" json:"follow_count,omitempty" gorm:"column:follow_count;type:int(10) unsigned;default:0;comment:关注数;"`
	FansCount     int       `form:"fans_count" json:"fans_count,omitempty" gorm:"column:fans_count;type:int(10) unsigned;default:0;comment:粉丝数;"`
	FavoriteCount int       `form:"favorite_count" json:"favorite_count,omitempty" gorm:"column:favorite_count;type:int(10) unsigned;default:0;comment:收藏数;"`
	CommentCount  int       `form:"comment_count" json:"comment_count,omitempty" gorm:"column:comment_count;type:int(11);size:11;default:0;comment:评论数;"`
	Status        int8      `form:"status" json:"status,omitempty" gorm:"column:status;type:tinyint(4);size:4;default:0;index:status;comment:用户状态：0正常 1禁用 2审核中 3审核拒绝 4审核忽略;"`
	Avatar        string    `form:"avatar" json:"avatar,omitempty" gorm:"column:avatar;type:varchar(255);size:255;comment:头像;"`
	Identity      string    `form:"identity" json:"identity,omitempty" gorm:"column:identity;type:char(18);size:18;comment:身份证号码;"`
	Realname      string    `form:"realname" json:"realname,omitempty" gorm:"column:realname;type:varchar(20);size:20;comment:身份证姓名;"`
	LoginAt       time.Time `form:"login_at" json:"login_at,omitempty" gorm:"column:login_at;type:datetime;comment:最后登录时间;"`
	CreatedAt     time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt     time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

// 这里是proto文件中的结构体，可以根据需要删除或者调整
//message User {
// int64 id = 1;
// string username = 2;
// string password = 3;
// string nickname = 4;
// string mobile = 5;
// string email = 6;
// string address = 7;
// string signature = 8;
// string last_login_ip = 9;
// string register_ip = 10;
// int32 doc_count = 11;
// int32 follow_count = 12;
// int32 fans_count = 13;
// int32 favorite_count = 14;
// int32 comment_count = 15;
// int32 status = 16;
// string avatar = 17;
// string identity = 18;
// string realname = 19;
// google.protobuf.Timestamp login_at = 20 [ (gogoproto.stdtime) = true ];
// google.protobuf.Timestamp created_at = 21 [ (gogoproto.stdtime) = true ];
// google.protobuf.Timestamp updated_at = 22 [ (gogoproto.stdtime) = true ];
//}

func (User) TableName() string {
	return tablePrefix + "user"
}

// CreateUser 创建User
// TODO: 创建成功之后，注意相关表统计字段数值的增减
func (m *DBModel) CreateUser(user *User, groupId int64) (err error) {
	user.Password, _ = unchained.MakePassword(user.Password, unchained.GetRandomString(4), "md5")

	sess := m.db.Begin()
	defer func() {
		if err != nil {
			sess.Rollback()
		} else {
			sess.Commit()
		}
	}()

	// 1. 添加用户
	err = sess.Create(user).Error
	if err != nil {
		m.logger.Error("CreateUser", zap.Error(err))
		return
	}

	// 2. 添加用户组
	group := &UserGroup{
		UserId:  user.Id,
		GroupId: groupId,
	}
	err = sess.Create(group).Error
	if err != nil {
		m.logger.Error("CreateUser", zap.Error(err))
		return
	}

	// 3. 添加用户统计
	err = sess.Model(&Group{}).Where("id = ?", groupId).Update("user_count", gorm.Expr("user_count + ?", 1)).Error
	if err != nil {
		m.logger.Error("CreateUser", zap.Error(err))
		return
	}

	return
}

// UpdateUserPassword 更新User密码
func (m *DBModel) UpdateUserPassword(id interface{}, newPassword string) (err error) {
	newPassword, _ = unchained.MakePassword(newPassword, unchained.GetRandomString(4), "md5")
	err = m.db.Model(&User{}).Where("id = ?", id).Update("password", newPassword).Error
	if err != nil {
		m.logger.Error("UpdateUserPassword", zap.Error(err))
	}
	return
}

// UpdateUser 更新User，如果需要更新指定字段，则请指定updateFields参数
func (m *DBModel) UpdateUser(user *User, updateFields ...string) (err error) {
	db := m.db.Model(user)

	updateFields = m.FilterValidFields(User{}.TableName(), updateFields...)
	if len(updateFields) > 0 { // 更新指定字段
		db = db.Select(updateFields)
	}

	err = db.Where("id = ?", user.Id).Updates(user).Error
	if err != nil {
		m.logger.Error("UpdateUser", zap.Error(err))
	}
	return
}

// GetUser 根据id获取User
func (m *DBModel) GetUser(id interface{}, fields ...string) (user User, err error) {
	db := m.db

	fields = m.FilterValidFields(User{}.TableName(), fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	err = db.Where("id = ?", id).First(&user).Error
	return
}

// GetUserByUsername(username string, fields ...string) 根据唯一索引获取User
func (m *DBModel) GetUserByUsername(username string, fields ...string) (user User, err error) {
	db := m.db

	fields = m.FilterValidFields(User{}.TableName(), fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	db = db.Where("username = ?", username)

	err = db.First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserByUsername", zap.Error(err))
		return
	}
	return
}

type OptionGetUserList struct {
	Page         int
	Size         int
	WithCount    bool                      // 是否返回总数
	Ids          []interface{}             // id列表
	SelectFields []string                  // 查询字段
	QueryRange   map[string][2]interface{} // map[field][]{min,max}
	QueryIn      map[string][]interface{}  // map[field][]{value1,value2,...}
	QueryLike    map[string][]interface{}  // map[field][]{value1,value2,...}
	Sort         []string
}

// GetUserList 获取User列表
func (m *DBModel) GetUserList(opt OptionGetUserList) (userList []User, total int64, err error) {
	db := m.db.Model(&User{})

	for field, rangeValue := range opt.QueryRange {
		fields := m.FilterValidFields(User{}.TableName(), field)
		if len(fields) == 0 {
			continue
		}
		if rangeValue[0] != nil {
			db = db.Where(fmt.Sprintf("%s >= ?", field), rangeValue[0])
		}
		if rangeValue[1] != nil {
			db = db.Where(fmt.Sprintf("%s <= ?", field), rangeValue[1])
		}
	}

	for field, values := range opt.QueryIn {
		fields := m.FilterValidFields(User{}.TableName(), field)
		if len(fields) == 0 {
			continue
		}
		db = db.Where(fmt.Sprintf("%s in (?)", field), values)
	}

	for field, values := range opt.QueryLike {
		fields := m.FilterValidFields(User{}.TableName(), field)
		if len(fields) == 0 {
			continue
		}
		db = db.Where(strings.TrimSuffix(fmt.Sprintf(strings.Join(make([]string, len(values)+1), "%s like ? or"), field), "or"), values...)
	}

	if len(opt.Ids) > 0 {
		db = db.Where("id in (?)", opt.Ids)
	}

	if opt.WithCount {
		err = db.Count(&total).Error
		if err != nil {
			m.logger.Error("GetUserList", zap.Error(err))
			return
		}
	}

	opt.SelectFields = m.FilterValidFields(User{}.TableName(), opt.SelectFields...)
	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	}

	if len(opt.Sort) > 0 {
		var sorts []string
		for _, sort := range opt.Sort {
			slice := strings.Split(sort, " ")
			if len(m.FilterValidFields(User{}.TableName(), slice[0])) == 0 {
				continue
			}

			if len(slice) == 2 {
				sorts = append(sorts, fmt.Sprintf("%s %s", slice[0], slice[1]))
			} else {
				sorts = append(sorts, fmt.Sprintf("%s desc", slice[0]))
			}
		}
		if len(sorts) > 0 {
			db = db.Order(strings.Join(sorts, ","))
		}
	}

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&userList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserList", zap.Error(err))
	}
	return
}

// DeleteUser 删除数据
// TODO: 删除数据之后，存在 user_id 的关联表，需要删除对应数据，同时相关表的统计数值，也要随着减少
func (m *DBModel) DeleteUser(ids []interface{}) (err error) {
	err = m.db.Where("id in (?)", ids).Delete(&User{}).Error
	if err != nil {
		m.logger.Error("DeleteUser", zap.Error(err))
	}
	return
}

type UserClaims struct {
	UserId int64
	UUID   string
	jwt.StandardClaims
}

// CreateUserJWTToken 生成用户JWT Token
func (m *DBModel) CreateUserJWTToken(userId int64) (string, error) {
	jwtCfg := m.GetConfigOfJWT()
	expireTime := time.Now().Add(time.Duration(jwtCfg.Duration) * 24 * time.Hour).Unix()
	claims := UserClaims{
		UserId: userId,
		UUID:   uuid.Must(uuid.NewV4()).String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime,
			Issuer:    "moredoc",
			IssuedAt:  time.Now().Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtCfg.Secret))
	return token, err
}

// CheckUserJWTToken 验证用户JWT token
func (m *DBModel) CheckUserJWTToken(token string) (*UserClaims, error) {
	jwtCfg := m.GetConfigOfJWT()
	tokenClaims, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*UserClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

func (m *DBModel) initUser() (err error) {
	// 如果不存在任意用户，则初始化一个用户作为管理员
	var existUser User
	m.db.Select("id").First(&existUser)
	if existUser.Id > 0 {
		return
	}

	// 初始化一个用户
	user := &User{Username: "admin", Password: "123456"}
	groupId := 1 // ID==1的用户组为管理员组
	err = m.CreateUser(user, int64(groupId))
	if err != nil {
		m.logger.Error("initUser", zap.Error(err))
	}
	return
}