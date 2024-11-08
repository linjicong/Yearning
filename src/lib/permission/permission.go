package permission

import (
	"Yearning-go/src/lib/vars"
	"Yearning-go/src/model"
	mapset "github.com/deckarep/golang-set/v2"
	"gorm.io/gorm"
)

// PermissionService 提供权限相关的服务
type PermissionService struct {
	db *gorm.DB
}

type Control struct {
	User     string
	Kind     int
	WorkId   string
	SourceId string
}

// NewPermissionService 创建一个新的权限服务实例
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CreatePermissionListFromGroups 从组ID列表创建一个合并的权限列表
func (service *PermissionService) CreatePermissionListFromGroups(groupIDs []string) *model.PermissionList {
	combinedPermissions := new(model.PermissionList)

	for _, groupID := range groupIDs {
		var coreRoleGroup model.CoreRoleGroup
		var permissions model.PermissionList

		// 查询数据库中的角色组
		if err := service.db.Where("group_id = ?", groupID).First(&coreRoleGroup).Error; err != nil {
			model.DefaultLogger.Errorf("Error fetching group with ID %s: %v", groupID, err)
			continue
		}

		// 将权限数据解码到权限列表中
		if err := coreRoleGroup.Permissions.UnmarshalToJSON(&permissions); err != nil {
			model.DefaultLogger.Errorf("Error unmarshalling permissions for group ID %s: %v", groupID, err)
			continue
		}

		// 合并权限列表
		appendPermissions(combinedPermissions, &permissions)
	}

	// 去除重复的权限
	removeDuplicatePermissions(combinedPermissions)

	return combinedPermissions
}

// Equal 检查用户对资源的权限
func (service *PermissionService) Equal(control *Control) bool {
	var grained model.CoreGrained
	var roleGroups []string

	// 查询数据库中的精细权限
	if err := service.db.Model(model.CoreGrained{}).Where("username = ?", control.User).First(&grained).Error; err != nil {
		model.DefaultLogger.Infof("Error fetching grained permissions for user %s: %v", control.User, err)
		return false
	}

	// 解码组信息
	if err := grained.Group.UnmarshalToJSON(&roleGroups); err != nil {
		model.DefaultLogger.Errorf("Error unmarshalling group information for user %s: %v", control.User, err)
		return false
	}

	// 获取用户规则集
	permissions := service.CreatePermissionListFromGroups(roleGroups)
	// 检查权限
	switch control.Kind {
	case vars.DDL:
		return mapset.NewSet[string](permissions.DDLSource...).Contains(control.SourceId)
	case vars.DML:
		return mapset.NewSet[string](permissions.DMLSource...).Contains(control.SourceId)
	case vars.QUERY:
		return mapset.NewSet[string](permissions.QuerySource...).Contains(control.SourceId)
	}

	return false
}

// appendPermissions 将源权限列表中的权限追加到目标权限列表中
func appendPermissions(target, source *model.PermissionList) {
	target.DDLSource = append(target.DDLSource, source.DDLSource...)
	target.DMLSource = append(target.DMLSource, source.DMLSource...)
	target.QuerySource = append(target.QuerySource, source.QuerySource...)
}

// removeDuplicatePermissions 移除权限列表中的重复项
func removeDuplicatePermissions(permissionList *model.PermissionList) {
	permissionList.DDLSource = mapset.NewSet[string](permissionList.DDLSource...).ToSlice()
	permissionList.DMLSource = mapset.NewSet[string](permissionList.DMLSource...).ToSlice()
	permissionList.QuerySource = mapset.NewSet[string](permissionList.QuerySource...).ToSlice()
}
