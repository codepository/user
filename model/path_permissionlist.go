package model

// Permissionlist Permissionlist
type Permissionlist struct {
	ID   int    `gorm:"primary_key" json:"id,omitempty"`
	Role string `json:"role,omitempty"`
	URL  string `json:"url,omitemtpy"`
}

// FindAllPermission FindAllPermission
func FindAllPermission(query interface{}) ([]*Permissionlist, error) {
	var result []*Permissionlist
	err := db.Where(query).Find(&result).Error
	return result, err
}

// FindAllPermissionByRoles FindAllPermissionByRoles
func FindAllPermissionByRoles(roles []string) ([]*Permissionlist, error) {
	var result []*Permissionlist
	err := db.Where("role in (?)", roles).Find(&result).Error
	return result, err
}
