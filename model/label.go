package model

import "errors"

type labelStatus int

// Label 标签
type Label struct {
	Model
	Name     string `json:"name"`
	Type     string `json:"type"`
	Describe string `json:"describe,omitempty"`
}

// FromMap FromMap
func (l *Label) FromMap(json map[string]interface{}) error {
	name, yes := json["name"].(string)
	if !yes {
		return errors.New("name 须为字符串")
	}
	l.Name = name
	type1, yes := json["type"].(string)
	if !yes {
		return errors.New("type 须为字符串")
	}
	l.Type = type1
	describe, yes := json["describe"].(string)
	if !yes {
		return errors.New("describe 须为字符串")
	}
	l.Describe = describe
	return nil

}

// SaveOrUpdate 不存在就保存
func (l *Label) SaveOrUpdate() error {
	//
	if len(l.Name) == 0 || len(l.Type) == 0 || len(l.Describe) == 0 {
		return errors.New("name、type、describe都不能为空;name是标签的名称,type是标签的类型,describe是标签的作用")
	}
	return db.Where(Label{Name: l.Name}).Assign(l).FirstOrCreate(l).Error
}

// GetLabels 获取所有标签
func GetLabels() ([]*Label, error) {
	var result []*Label
	err := db.Find(&result).Error
	return result, err
}
