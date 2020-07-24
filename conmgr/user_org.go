package conmgr

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/codepository/user/model"
)

// FindOrgidsByUserid 根据用户id查询用户可以查看的行业
func FindOrgidsByUserid(c *model.Container) error {
	if c.Body.UserID == 0 {
		return errors.New("user_id不能为空")
	}
	result, err := model.FindAllOrg(map[string]interface{}{"userid": c.Body.UserID}, "orgid")
	if err != nil {
		return err
	}

	var orgids []int
	for _, o := range result {
		orgids = append(orgids, o.Orgid)
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, orgids)
	return nil
}

// FindUserByOrgid 根据行业orgid查询用户
func FindUserByOrgid(c *model.Container) error {
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New(`参数格式:{"body":{"data":[{"orgid": 2}],"metrics":"userid,username"}}},orgid为数字,metrics为返回字段`)
	}
	if len(c.Body.Metrics) == 0 {
		c.Body.Metrics = "*"
	}
	par, ok := c.Body.Data[0].(map[string]interface{})
	if len(par) == 0 {
		return errors.New(`参数格式:{"body":{"data":[{"orgid": 2}],"metrics":"userid,username"}}},orgid为数字,metrics为返回字段`)
	}
	if !ok {
		return errors.New(`参数格式:{"body":{"data":[{"orgid": 2}],"metrics":"userid,username"}}},orgid为数字,metrics为返回字段`)
	}
	result, err := model.FindAllOrg(par, c.Body.Metrics)
	if err != nil {
		return err
	}
	c.Body.Data = c.Body.Data[:0]
	c.Body.Data = append(c.Body.Data, result)
	return nil
}

// DelUserOrgByID 删除
func DelUserOrgByID(c *model.Container) error {
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New(`参数格式:{"body":"data":[[2,3]]},2,3为要删除字段的id`)
	}
	par, ok := c.Body.Data[0].([]interface{})
	if len(par) == 0 {
		return errors.New(`参数格式:{"body":"data":[[2,3]]},2,3为要删除字段的id`)
	}
	if !ok {
		return errors.New(`参数格式:{"body":"data":[[2,3]]},2,3为要删除字段的id`)
	}
	err := model.DelUserOrg(par)
	if err != nil {

		return err
	}
	return nil
}

// SaveUserOrg SaveUserOrg
func SaveUserOrg(c *model.Container) error {
	if c.Body.Data == nil || len(c.Body.Data) == 0 {
		return errors.New(`参数格式:{"body":"data":[{"orgid":2,"org":"食品饮料","userid":334,"username":"张三"}]}`)
	}
	par, ok := c.Body.Data[0].(map[string]interface{})
	fmt.Println(reflect.TypeOf(c.Body.Data[0]))
	if !ok {
		return errors.New(`参数格式:{"body":"data":[{"orgid":2,"org":"食品饮料","userid":334,"username":"张三"}]}`)
	}
	var entity model.UserOrg
	entity.FromMap(par)
	err := entity.Save()
	if err != nil {
		return err
	}
	c.Body.Data = nil

	return nil
}
