package model

// FznewsRecord 纪录
type FznewsRecord struct {
	Model
	Data string `json:"data" gorm:"size:1024"`
	Type string `json:"type"`
	Flag uint8  `json:"flag"` // 0失败，1成功
	Err  string `json:"err"`
}

// Save 直接保存
func (r *FznewsRecord) Save() error {
	return db.Create(r).Error
}
