package model

type Book struct {
	ID         int    `json:"id" gorm:"primaryKey;AUTO_INCREMENT"`
	Title      string `json:"title"`
	Format     string `json:"format"`
	Annotation string `json:"annotation"`
	Author     string `json:"author"`
	Filepath   string `json:"filepath"`
	Chapters   uint   `json:"chapters"`
	Pages      uint   `json:"pages"`
	CreatedAt  int64  `json:"created_at"`
	UserId     int    `json:"user_id"`
}

type User struct {
	ID       int    `json:"id" gorm:"primaryKey;AUTO_INCREMENT"`
	Login    string `gorm:"unique" json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   int    `json:"-"`
	Role     Role   `gorm:"foreignKey:RoleID" json:"role"`
}

type Role struct {
	ID       int    `json:"id" gorm:"primaryKey;AUTO_INCREMENT"`
	RoleName string `json:"role_name" gorm:"unique"`
}
type ReadingProgress struct {
	ID          int   `json:"id" gorm:"primaryKey;AUTO_INCREMENT"`
	UserID      int   `json:"user_id" gorm:"not null"`
	BookID      int   `json:"book_id" gorm:"not null"`
	CurrentPage int   `json:"current_page" gorm:"default:1"`
	LastReadAt  int64 `json:"last_read_at"`
	User        User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	Book        Book  `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE" json:"book"`
}
