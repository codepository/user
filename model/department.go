package model

var wxDepartmentTableName = "weixin_leave_department"

// WxDepartment 微信部门
type WxDepartment struct {
	Model
	Name     string `json:"name"`
	Parentid int    `json:"parentid"`
	Order    int    `json:"order,omitempty"`
	// 0删除，1有效
	St int8 `json:"st"`
}

// TreeNode 树形节点
type TreeNode struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Parentid int         `json:"parentid"`
	Children []*TreeNode `json:"children"`
}

// FindAllWxDepartment 查询所有部门
func FindAllWxDepartment() ([]*WxDepartment, error) {
	var result []*WxDepartment
	err := wxdb.Table(wxDepartmentTableName).Where("st=1").Find(&result).Error
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
			Name:     d.Name,
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
