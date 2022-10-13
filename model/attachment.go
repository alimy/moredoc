package model

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Attachment struct {
	Id         int64     `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:附件 id;"`
	Hash       string    `form:"hash" json:"hash,omitempty" gorm:"column:hash;type:char(32);size:32;index:hash;comment:文件MD5;"`
	UserId     int64     `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20) unsigned;default:0;index:user_id;comment:用户 id;"`
	TypeId     int64     `form:"type_id" json:"type_id,omitempty" gorm:"column:type_id;type:bigint(20) unsigned;default:0;comment:类型数据ID，对应与用户头像时，则为用户id，对应为文档时，则为文档ID;"`
	Type       int       `form:"type" json:"type,omitempty" gorm:"column:type;type:smallint(5) unsigned;default:0;comment:附件类型(0 头像，1 文档，2 文章附件 ...);"`
	IsApproved int8      `form:"is_approved" json:"is_approved,omitempty" gorm:"column:is_approved;type:tinyint(3) unsigned;default:1;comment:是否合法;"`
	Path       string    `form:"path" json:"path,omitempty" gorm:"column:path;type:varchar(255);size:255;comment:文件存储路径;"`
	Name       string    `form:"name" json:"name,omitempty" gorm:"column:name;type:varchar(255);size:255;comment:文件原名称;"`
	Size       int64     `form:"size" json:"size,omitempty" gorm:"column:size;type:bigint(20) unsigned;default:0;comment:文件大小;"`
	Width      int64     `form:"width" json:"width,omitempty" gorm:"column:width;type:bigint(20) unsigned;default:0;comment:宽度;"`
	Height     int64     `form:"height" json:"height,omitempty" gorm:"column:height;type:bigint(20) unsigned;default:0;comment:高度;"`
	Ext        string    `form:"ext" json:"ext,omitempty" gorm:"column:ext;type:varchar(32);size:32;comment:文件类型，如 .pdf 。统一处理成小写;"`
	Ip         string    `form:"ip" json:"ip,omitempty" gorm:"column:ip;type:varchar(16);size:16;comment:上传文档的用户IP地址;"`
	CreatedAt  time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt  time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

// 这里是proto文件中的结构体，可以根据需要删除或者调整
//message Attachment {
// int64 id = 1;
// string hash = 2;
// int64 user_id = 3;
// int64 type_id = 4;
// int32 type = 5;
// int32 is_approved = 6;
// string path = 7;
// string name = 8;
// int64 size = 9;
// int64 width = 10;
// int64 height = 11;
// string ext = 12;
// string ip = 13;
// google.protobuf.Timestamp created_at = 14 [ (gogoproto.stdtime) = true ];
// google.protobuf.Timestamp updated_at = 15 [ (gogoproto.stdtime) = true ];
//}

func (Attachment) TableName() string {
	return tablePrefix + "attachment"
}

// CreateAttachment 创建Attachment
// TODO: 创建成功之后，注意相关表统计字段数值的增减
func (m *DBModel) CreateAttachment(attachment *Attachment) (err error) {
	err = m.db.Create(attachment).Error
	if err != nil {
		m.logger.Error("CreateAttachment", zap.Error(err))
		return
	}
	return
}

// UpdateAttachment 更新Attachment，如果需要更新指定字段，则请指定updateFields参数
func (m *DBModel) UpdateAttachment(attachment *Attachment, updateFields ...string) (err error) {
	db := m.db.Model(attachment)

	updateFields = m.FilterValidFields(Attachment{}.TableName(), updateFields...)
	if len(updateFields) > 0 { // 更新指定字段
		db = db.Select(updateFields)
	}

	err = db.Where("id = ?", attachment.Id).Updates(attachment).Error
	if err != nil {
		m.logger.Error("UpdateAttachment", zap.Error(err))
	}
	return
}

// GetAttachment 根据id获取Attachment
func (m *DBModel) GetAttachment(id interface{}, fields ...string) (attachment Attachment, err error) {
	db := m.db

	fields = m.FilterValidFields(Attachment{}.TableName(), fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	err = db.Where("id = ?", id).First(&attachment).Error
	return
}

type OptionGetAttachmentList struct {
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

// GetAttachmentList 获取Attachment列表
func (m *DBModel) GetAttachmentList(opt OptionGetAttachmentList) (attachmentList []Attachment, total int64, err error) {
	db := m.db.Model(&Attachment{})

	for field, rangeValue := range opt.QueryRange {
		fields := m.FilterValidFields(Attachment{}.TableName(), field)
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
		fields := m.FilterValidFields(Attachment{}.TableName(), field)
		if len(fields) == 0 {
			continue
		}
		db = db.Where(fmt.Sprintf("%s in (?)", field), values)
	}

	for field, values := range opt.QueryLike {
		fields := m.FilterValidFields(Attachment{}.TableName(), field)
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
			m.logger.Error("GetAttachmentList", zap.Error(err))
			return
		}
	}

	opt.SelectFields = m.FilterValidFields(Attachment{}.TableName(), opt.SelectFields...)
	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	}

	if len(opt.Sort) > 0 {
		var sorts []string
		for _, sort := range opt.Sort {
			slice := strings.Split(sort, " ")
			if len(m.FilterValidFields(Attachment{}.TableName(), slice[0])) == 0 {
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

	err = db.Find(&attachmentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetAttachmentList", zap.Error(err))
	}
	return
}

// DeleteAttachment 删除数据
// TODO: 删除数据之后，存在 attachment_id 的关联表，需要删除对应数据，同时相关表的统计数值，也要随着减少
func (m *DBModel) DeleteAttachment(ids []interface{}) (err error) {
	err = m.db.Where("id in (?)", ids).Delete(&Attachment{}).Error
	if err != nil {
		m.logger.Error("DeleteAttachment", zap.Error(err))
	}
	return
}