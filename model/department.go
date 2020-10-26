package model

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var wxDepartmentTableName = "weixin_leave_department"

// WxDepartment 微信部门
type WxDepartment struct {
	Model
	Name     string `json:"name"`
	Parentid int    `json:"parentid"`
	Leader   string `json:"leader,omitempty"`
	Order    int    `json:"order,omitempty"`
	// 部门层级
	Level int `json:"level,omitempty"`
	// 0删除，1有效
	St int8 `json:"st"`
}

// TreeNode 树形节点
type TreeNode struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Parentid   int    `json:"parentid"`
	Leader     string `json:"leader,omitempty"`
	LeaderType string `json:"leader_type,omitempty"`
	// 部门层级
	Level    int         `json:"level,omitempty"`
	Children []*TreeNode `json:"children"`
}

// FindAllDepartment 查询所有部门
func FindAllDepartment(query interface{}, values ...interface{}) ([]*WxDepartment, error) {
	var result []*WxDepartment
	err := wxdb.Table(wxDepartmentTableName+" d").Select("d.*").
		Where("d.st=1").Where(query, values...).
		Find(&result).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return result, nil
}

// FindDepartmentByID FindDepartmentByID
func FindDepartmentByID(id int) (*WxDepartment, error) {
	depts, err := FindAllDepartment("id=?", id)
	if err != nil {
		return nil, err
	}
	if len(depts) == 0 {
		return nil, errors.New("部门不存在")
	}
	return depts[0], nil
}

// SetDepartmentLevelByID 根据部门ID设置部门level，并返回level值
func SetDepartmentLevelByID(id int) (int, error) {
	var err error
	dept, err := FindDepartmentByID(id)
	if err != nil {
		return 0, err
	}
	level := dept.Level
	if level != 0 {
		return level, nil
	}
	var parent = dept
	// 迭代找父节点
	for {
		// 如果父节点为0，则结束
		if parent.Parentid == 0 {
			break
		}
		if parent.Level != 0 {
			level = level + parent.Level
			break
		} else {
			level++
		}
		// 查询下一个父节点
		parent, err = FindDepartmentByID(parent.Parentid)
		if err != nil {
			return 0, err
		}
	}
	// 更新部门level
	dept.Level = level
	err = dept.UpdateLevel()
	if err != nil {
		return 0, err
	}
	return level, nil

}

// UpdateLevel 更新部门层级
func (d *WxDepartment) UpdateLevel() error {
	return wxdb.Table(wxDepartmentTableName).Model(d).UpdateColumn("level", d.Level).Error
}

// FindAllWxDepartment 查询所有部门和部门领导
func FindAllWxDepartment() ([]*WxDepartment, error) {
	var result []*WxDepartment
	err := wxdb.Table(wxDepartmentTableName + " d").Select("d.*").
		Where("d.st=1").
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	// 查询fznews_leadership
	leaders, err := FindFznewsLeadership("")
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if len(leaders) > 0 {
		lmap := make(map[int]string)
		for _, l := range leaders {
			if len(lmap[l.DepartmentID]) == 0 {
				lmap[l.DepartmentID] = l.Leader
			} else {
				lmap[l.DepartmentID] = lmap[l.DepartmentID] + "," + l.Leader
			}

		}
		for _, r := range result {
			r.Leader = lmap[r.ID]
		}
	}

	return result, err
}

// TransformWxDepartment2Tree 部门列表转换成树形结构
// 1、所有节点压入栈all[]
// 2、all依次取出元素，将当前节点的父节点的放入不重复集合 parent[]，剩下的为当前叶节点
// 3、all依次取出元素，并在parent[]找寻父子点，并并入父节点，最后将parent[]中的元素压入all[]
// 重复2、3步骤，直到parent[]中只有一个元素
func TransformWxDepartment2Tree(list []*WxDepartment) []*TreeNode {
	if list == nil || len(list) == 0 {
		return nil
	}
	var result []*TreeNode
	// 父节点相同的归到一起
	all := make(map[int]*TreeNode, len(list))
	for _, d := range list {
		all[d.ID] = &TreeNode{
			ID: d.ID, Parentid: d.Parentid,
			Leader:   d.Leader,
			Name:     d.Name,
			Level:    d.Level,
			Children: []*TreeNode{}}
	}
	for {
		// 父节点id
		parentids := make(map[int]int)
		for _, v := range all {
			parentids[v.Parentid] = v.Parentid
		}
		if len(parentids) == 0 {
			break
		}
		parent := make(map[int]*TreeNode, len(parentids))
		// 找到父节点
		for _, p := range parentids {

			if all[p] != nil {
				parent[p] = all[p]
				delete(all, p)
			}
		}
		if len(parent) == 0 {
			// 返回结果
			break
		}
		// 将子节点放入父节点
		for k, v := range all {
			if parent[v.Parentid] != nil {
				parent[v.Parentid].Children = append(parent[v.Parentid].Children, v)
				delete(all, k)
			}
		}
		// 未找到父节点的子节点，放到结果中
		if len(all) != 0 {
			for p, v := range all {
				result = append(result, v)
				delete(all, p)
			}
		}
		all = parent

	}
	for _, v := range all {
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}
