package service

import (
	"context"

	"api-service/models"
	"api-service/repository"
)

// MenuService defines operations for menu management
type MenuService interface {
	GetMenuTreeByUserID(ctx context.Context, userID int) ([]*models.AdminMenu, error)
	GetAllMenuTree(ctx context.Context) ([]*models.AdminMenu, error)
	CreateMenu(ctx context.Context, menu *models.DBAdminMenu) (int, error)
	UpdateMenu(ctx context.Context, menu *models.DBAdminMenu) error
	DeleteMenu(ctx context.Context, id int) error
	CreateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) (int, error)
	UpdateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) error
	DeleteMenuAuth(ctx context.Context, id int) error
}

type menuService struct {
	menuRepo repository.MenuRepository
}

// NewMenuService creates an instance of MenuService
func NewMenuService(menuRepo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: menuRepo}
}

// buildMenuTree constructs a hierarchical menu tree from flat menus and auths
func buildMenuTree(flatMenus []*models.DBAdminMenu, auths []*models.DBAdminMenuAuth) []*models.AdminMenu {
	authMap := make(map[int][]models.AdminMenuAuthItem)
	for _, auth := range auths {
		authMap[auth.MenuID] = append(authMap[auth.MenuID], models.AdminMenuAuthItem{
			ID:       auth.ID,
			MenuID:   auth.MenuID,
			Title:    auth.Title,
			AuthMark: auth.AuthMark,
		})
	}

	menuMap := make(map[int]*models.AdminMenu)
	var rootMenus []*models.AdminMenu

	for _, flat := range flatMenus {
		menu := &models.AdminMenu{
			ID:        flat.ID,
			ParentID:  flat.ParentID,
			Name:      flat.Name,
			Path:      flat.Path,
			Component: flat.Component,
			Meta: models.AdminMenuMeta{
				Title:      flat.Title,
				Icon:       flat.Icon,
				IsHide:     flat.IsHide,
				KeepAlive:  flat.KeepAlive,
				IsHideTab:  flat.IsHideTab,
				IsFullPage: flat.IsFullPage,
				FixedTab:   flat.FixedTab,
				AuthList:   authMap[flat.ID],
			},
			Children: make([]*models.AdminMenu, 0),
		}
		menuMap[flat.ID] = menu
	}

	for _, flat := range flatMenus {
		menu, exists := menuMap[flat.ID]
		if !exists {
			continue
		}
		if flat.ParentID == 0 {
			rootMenus = append(rootMenus, menu)
		} else {
			if parent, exists := menuMap[flat.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return rootMenus
}

func (s *menuService) GetMenuTreeByUserID(ctx context.Context, userID int) ([]*models.AdminMenu, error) {
	flatMenus, err := s.menuRepo.GetMenusByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	auths, err := s.menuRepo.GetMenuAuthsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(flatMenus, auths), nil
}

func (s *menuService) GetAllMenuTree(ctx context.Context) ([]*models.AdminMenu, error) {
	flatMenus, err := s.menuRepo.GetAllMenus(ctx)
	if err != nil {
		return nil, err
	}
	auths, err := s.menuRepo.GetAllMenuAuths(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(flatMenus, auths), nil
}

func (s *menuService) CreateMenu(ctx context.Context, menu *models.DBAdminMenu) (int, error) {
	return s.menuRepo.CreateMenu(ctx, menu)
}

func (s *menuService) UpdateMenu(ctx context.Context, menu *models.DBAdminMenu) error {
	return s.menuRepo.UpdateMenu(ctx, menu)
}

func (s *menuService) DeleteMenu(ctx context.Context, id int) error {
	return s.menuRepo.DeleteMenu(ctx, id)
}

func (s *menuService) CreateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) (int, error) {
	return s.menuRepo.CreateMenuAuth(ctx, auth)
}

func (s *menuService) UpdateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) error {
	return s.menuRepo.UpdateMenuAuth(ctx, auth)
}

func (s *menuService) DeleteMenuAuth(ctx context.Context, id int) error {
	return s.menuRepo.DeleteMenuAuth(ctx, id)
}
