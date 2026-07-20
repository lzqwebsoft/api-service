package service

import (
	"context"

	"api-service/models"
	"api-service/repository"
)

// RoleService defines business operations for role management
type RoleService interface {
	ListRoles(ctx context.Context, roleName, roleCode string, page, size int) ([]*models.AdminRole, int, error)
	CreateRole(ctx context.Context, role *models.AdminRole) (int, error)
	UpdateRole(ctx context.Context, role *models.AdminRole) error
	DeleteRole(ctx context.Context, id int) error
	GetRoleMenuIDs(ctx context.Context, roleID int) ([]int, error)
	SetRoleMenus(ctx context.Context, roleID int, menuIDs []int) error
}

type roleService struct {
	roleRepo repository.RoleRepository
}

// NewRoleService creates an instance of RoleService
func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) ListRoles(ctx context.Context, roleName, roleCode string, page, size int) ([]*models.AdminRole, int, error) {
	return s.roleRepo.ListRoles(ctx, roleName, roleCode, page, size)
}

func (s *roleService) CreateRole(ctx context.Context, role *models.AdminRole) (int, error) {
	return s.roleRepo.CreateRole(ctx, role)
}

func (s *roleService) UpdateRole(ctx context.Context, role *models.AdminRole) error {
	return s.roleRepo.UpdateRole(ctx, role)
}

func (s *roleService) DeleteRole(ctx context.Context, id int) error {
	return s.roleRepo.DeleteRole(ctx, id)
}

func (s *roleService) GetRoleMenuIDs(ctx context.Context, roleID int) ([]int, error) {
	return s.roleRepo.GetRoleMenuIDs(ctx, roleID)
}

func (s *roleService) SetRoleMenus(ctx context.Context, roleID int, menuIDs []int) error {
	return s.roleRepo.SetRoleMenus(ctx, roleID, menuIDs)
}
