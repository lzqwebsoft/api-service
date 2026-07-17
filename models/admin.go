package models

import "time"

// AdminUser represents the database record of an administrator user
type AdminUser struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Omitted from JSON serialization for security
	CreatedAt    time.Time `json:"created_at"`
}

// AdminSession represents a login session token and its expiration rules
type AdminSession struct {
	ID               int       `json:"id"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	UserID           int       `json:"user_id"`
	Username         string    `json:"username"` // Retrieved via JOIN from admin_users
	AccessExpiresAt  int64     `json:"access_expires_at"`
	RefreshExpiresAt int64     `json:"refresh_expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// AdminLoginResult represents the login output containing tokens and their expiry times
type AdminLoginResult struct {
	Token            string    `json:"token"`
	RefreshToken     string    `json:"refreshToken"`
	ExpiresAt        time.Time `json:"expiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

// DBAdminMenu represents the flat menu structure in the database
type DBAdminMenu struct {
	ID         int
	ParentID   int
	Name       string
	Path       string
	Component  string
	Title      string
	Icon       string
	IsHide     bool
	KeepAlive  bool
	IsHideTab  bool
	IsFullPage bool
	FixedTab   bool
	SortOrder  int
}

// DBAdminMenuAuth represents a menu button permission in the database
type DBAdminMenuAuth struct {
	MenuID   int
	Title    string
	AuthMark string
}

// AdminMenu represents the dynamic route/menu node returned to the frontend
type AdminMenu struct {
	ID        int            `json:"id"`
	ParentID  int            `json:"parentId,omitempty"`
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Component string         `json:"component,omitempty"`
	Meta      AdminMenuMeta  `json:"meta"`
	Children  []*AdminMenu   `json:"children,omitempty"`
}

// AdminMenuMeta holds meta properties for AdminMenu
type AdminMenuMeta struct {
	Title      string              `json:"title"`
	Icon       string              `json:"icon,omitempty"`
	IsHide     bool                `json:"isHide,omitempty"`
	KeepAlive  bool                `json:"keepAlive,omitempty"`
	IsHideTab  bool                `json:"isHideTab,omitempty"`
	IsFullPage bool                `json:"isFullPage,omitempty"`
	FixedTab   bool                `json:"fixedTab,omitempty"`
	AuthList   []AdminMenuAuthItem `json:"authList,omitempty"`
}

// AdminMenuAuthItem holds individual button authority items
type AdminMenuAuthItem struct {
	Title    string `json:"title"`
	AuthMark string `json:"authMark"`
}
