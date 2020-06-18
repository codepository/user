package model

import "errors"

// FznewsLeadership 设置用户分管的部门
type FznewsLeadership struct {
	Model
	UserID         int    `json:"user_id"`
	DepartmentID   int    `json:"department_id"`
	DepartmentName string `json:"department_name"`
}

// SaveOrUpdate 不存在保存，存在就更新
func (f *FznewsLeadership) SaveOrUpdate() error {
	if f.UserID == 0 || f.DepartmentID == 0 {
		return errors.New("用户ID和部门ID不能为空")
	}
	return db.Where(FznewsLeadership{UserID: f.UserID, DepartmentID: f.DepartmentID}).Assign(*f).FirstOrCreate(f).Error
}

// DelFznewsLeadership 删除
func DelFznewsLeadership(query interface{}) error {
	if query == nil {
		return errors.New("参数不能为空")
	}
	return db.Where(query).Delete(&FznewsLeadership{}).Error
}

// FindFznewsLeadership FindFznewsLeadership
func FindFznewsLeadership(query interface{}) ([]*FznewsLeadership, error) {
	var result []*FznewsLeadership
	err := db.Table("fznews_leadership").Select("fznews_leadership.id,fznews_leadership.user_id,fznews_leadership.department_id,fznews_leadership.createTime,d.name as department_name").
		Joins("join weixin_leave_department d on d.id=department_id").
		Where(query).Find(&result).Error
	return result, err
}
