package conmgr

import (
	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
)

// AddNewLabel 添加标签
func AddNewLabel(c *model.Container) error {

	err := service.AddNewLabel(c)
	if err != nil {
		return err
	}
	result, err := model.GetLabels()
	if err != nil {
		return err
	}
	Conmgr.cacheMap[labelsCache] = result
	return nil
}

// FindAllLabel 查询所有标签
func FindAllLabel(c *model.Container) error {
	result, err := GetAllLabel()
	if err != nil {
		return err
	}
	c.Body.Data = append(c.Body.Data, result)
	return nil

}

// GetAllLabel GetAllLabel
func GetAllLabel() ([]*model.WeixinOauserTag, error) {
	var result []*model.WeixinOauserTag
	var err error
	data := Conmgr.cacheMap[labelsCache]
	if data == nil {
		result, err = service.FindAllLabel()
		if err != nil {
			return nil, err
		}
		Conmgr.cacheMap[labelsCache] = result
	} else {
		result = data.([]*model.WeixinOauserTag)
	}
	return result, nil
}

// GetLabelNamesByIds 获取标签
func GetLabelNamesByIds(ids []int) ([]string, error) {
	labels, err := GetAllLabel()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, id := range ids {
		for _, label := range labels {
			if label.ID == id {
				result = append(result, label.TagName)
				break
			}
		}
	}
	return result, nil
}
