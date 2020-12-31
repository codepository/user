package conmgr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codepository/user/model"
	"github.com/codepository/user/service"
	"github.com/mumushuiding/util"
)

// FindLabel 根据条件从weixin_oauser_tag查询数据
func FindLabel(c *model.Container) error {
	errstr := `参数格式:{"body":{"params":{"field":"id,tagName,type,describe","type":"标签类型","tagName":"标签名","tagNameLike":"根据标签名模糊查询"}}}`
	if len(c.Body.Params) == 0 {
		return fmt.Errorf(errstr)
	}
	var sqlbuff strings.Builder
	if c.Body.Params["tagNameLike"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and tagName like '%s' ", "%"+c.Body.Params["tagNameLike"].(string)+"%"))
	}
	if c.Body.Params["tagName"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and tagName='%s' ", c.Body.Params["tagName"].(string)))
	}
	if c.Body.Params["type"] != nil {
		sqlbuff.WriteString(fmt.Sprintf("and type='%s' ", c.Body.Params["type"].(string)))
	}
	if sqlbuff.Len() == 0 {
		return fmt.Errorf(errstr)
	}
	field := ""
	if c.Body.Params["field"] != nil {
		field = c.Body.Params["field"].(string)
	}
	datas, err := model.FindAllTags(field, sqlbuff.String()[4:])
	if err != nil {
		return fmt.Errorf("查询用户标签:%s", err.Error())
	}
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

// FindUserLabel 查询用户标签
func FindUserLabel(c *model.Container) error {
	errstr := `参数格式:{"body":{"params":{"userid":3,"labeltype":"考核组"}} labeltype是用户标签类别`
	if len(c.Body.Params) == 0 {
		return errors.New(errstr)
	}
	userid, err := util.Interface2Int(c.Body.Params["userid"])
	if err != nil {
		return err
	}
	var querybuffer strings.Builder
	if userid != 0 {
		querybuffer.WriteString(fmt.Sprintf(" and id in (select tagId from weixin_oauser_taguser where uId=%d)", userid))
	}
	labeltype := c.Body.Params["labeltype"]
	if labeltype != nil && len(labeltype.(string)) > 0 {
		querybuffer.WriteString(" and type='" + labeltype.(string) + "'")
	}
	if querybuffer.Len() == 0 {
		return errors.New(errstr)
	}
	datas, err := model.FindAllTags("", querybuffer.String()[5:])
	if err != nil {
		return fmt.Errorf("查询标签错误:%s", err.Error())
	}
	c.Body.Data = append(c.Body.Data, datas)
	return nil
}

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
