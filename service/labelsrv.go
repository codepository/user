package service

import (
	"errors"

	"github.com/codepository/user/model"
)

// AddNewLabel 添加新标签
func AddNewLabel(c *model.Container) error {
	//
	// 检查参数
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New("参数类型{'header':{'token':''},'body':{'data':[{'name':'','type':'','describe':''},{}]}}")
	}
	// 检查类型

	// 保存
	for _, v := range c.Body.Data {
		json := v.(map[string]interface{})
		l := &model.Label{}
		l.FromMap(json)
		err := l.SaveOrUpdate()
		return err
	}
	c.Body.Data = nil
	return nil
}
