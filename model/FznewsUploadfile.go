package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

// FznewsUploadfileTable 对应表格名
var FznewsUploadfileTable = "fznews_uploadfile"

// FznewsUploadfile 文件存储类型
type FznewsUploadfile struct {
	Model
	// Filetype 文件类型
	Filetype string `json:"filetype,omitempty"`
	// Filename 文件名
	Filename string `json:"filename,omitempty"`
	// Blob 文件二进制数据
	Blob     []byte `json:"blob,omitempty"`
	Username string `json:"username,omitempty"`
	UID      int    `json:"uid"`
}

// Create 生成文件
func (f *FznewsUploadfile) Create() error {
	f.CreateTime = time.Now()
	return wxdb.Create(f).Error
}

// DelUploafileByID 删除文件
func DelUploafileByID(id interface{}) error {
	return wxdb.Where("id=?", id).Delete(FznewsUploadfile{}).Error
}

// FindUploadfileByID 根据ID查询
func FindUploadfileByID(id interface{}) (*FznewsUploadfile, error) {
	data := &FznewsUploadfile{}
	err := wxdb.Table(FznewsUploadfileTable).Where("id=?", id).Limit(1).Find(&data).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return data, err
}

// FindUploadfiles 查询文件
func FindUploadfiles(fields, order string, limit, offset int, query interface{}) ([]*FznewsUploadfile, error) {
	if len(fields) == 0 {
		fields = "*"
	}
	if len(order) == 0 {
		order = "createTime desc"
	}
	if limit == 0 {
		limit = 10
	}
	var data []*FznewsUploadfile
	err := wxdb.Table(FznewsUploadfileTable).
		Select(fields).
		Where(query).Order(order).Limit(limit).Offset(offset).
		Find(&data).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return make([]*FznewsUploadfile, 0), nil
	}
	return data, err
}
