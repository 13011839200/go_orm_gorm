package model_defintion

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// -------------------------- 定义模型（含一对多关系标签） --------------------------
type User struct {
	gorm.Model
	Username     string `gorm:"size:50;not null;uniqueIndex;comment:用户名"`        // 唯一索引，非空
	Email        string `gorm:"size:100;uniqueIndex;comment:用户邮箱"`               // 唯一索引
	PasswordHash string `gorm:"size:255;not null;comment:密码哈希"`                  // 密码加密存储
	PostCount    int64  `gorm:"default:0;comment:用户发布的文章数量"`                     // 文章数统计字段
	Posts        []Post `gorm:"foreignKey:UserID;references:ID;comment:用户发布的文章"` // 一对多关联 Post
}

// Post 模型（文章）：
type Post struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null;comment:文章标题"`                // 非空
	Content       string    `gorm:"type:text;not null;comment:文章内容"`               // 长文本
	UserID        uint      `gorm:"not null;index;comment:所属用户ID"`                 // 外键（关联 User.ID），索引提升查询效率
	CommentStatus string    `gorm:"size:20;default:'无评论';comment:评论状态（有评论/无评论）"`   // 评论状态
	Comments      []Comment `gorm:"foreignKey:PostID;references:ID;comment:文章的评论"` // 一对多关联 Comment
}

// -------------------------- Post 钩子：创建时更新用户文章数 --------------------------
func (p *Post) AfterCreate(tx *gorm.DB) error {
	// 1. 更新用户的 PostCount 字段（原子递增，避免并发问题）
	if err := tx.Model(&User{}).Where("id = ?", p.UserID).UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("更新用户文章数失败：%v", err)
	}
	// 2. 文章创建时默认无评论，但如果有初始化评论可额外处理（此处无需）
	return nil
}

type Comment struct {
	gorm.Model        // 内置主键/时间字段
	Content    string `gorm:"type:text;not null;comment:评论内容"` // 非空
	PostID     uint   `gorm:"not null;index;comment:所属文章ID"`   // 外键（关联 Post.ID），索引
	UserID     uint   `gorm:"not null;index;comment:评论用户ID"`   // 扩展：记录评论者ID（关联 User.ID）
}

// -------------------------- Comment 钩子：创建时更新文章评论状态 --------------------------
func (c *Comment) AfterCreate(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		return fmt.Errorf("统计文章评论数失败：%v", err)
	}
	commentStatus := "无评论"
	if count > 0 {
		commentStatus = "有评论"
	}
	if err := tx.Model(&Post{}).Where(c.PostID).UpdateColumn("comment_status", commentStatus).Error; err != nil {
		return fmt.Errorf("更新文章评论状态失败：%v", err)
	}
	return nil
}

// -------------------------- Comment 钩子：删除时检查评论数并更新文章状态 --------------------------
// AfterDelete Comment 模型的删除后钩子
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	// 1. 统计当前文章的剩余评论数
	var commentCount int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&commentCount).Error; err != nil {
		return fmt.Errorf("统计文章评论数失败：%v", err)
	}

	// 2. 根据评论数更新文章的 CommentStatus
	status := "有评论"
	if commentCount == 0 {
		status = "无评论"
	}
	if err := tx.Model(&Post{}).Where("id = ?", c.PostID).Update("comment_status", status).Error; err != nil {
		return fmt.Errorf("更新文章评论状态失败：%v", err)
	}

	return nil
}

// -------------------------- 初始化数据库连接 --------------------------
func InitDB() *gorm.DB {
	// 数据库连接信息（根据实际环境修改）
	dsn := "root:88888888@tcp(127.0.0.1:3306)/go_store?charset=utf8mb4&parseTime=True&loc=Local"

	// 自定义日志配置（打印 SQL 语句，方便调试）
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到控制台
		logger.Config{
			SlowThreshold: time.Second, // 慢查询阈值
			LogLevel:      logger.Info, // 日志级别：Info 显示 SQL 语句
			Colorful:      true,        // 彩色输出
		},
	)

	// 连接 MySQL 并初始化 GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger, // 启用日志
	})
	if err != nil {
		log.Fatalf("数据库连接失败：%v", err)
	}

	// -------------------------- 3. 自动迁移（创建/更新表结构） --------------------------
	// 按依赖顺序迁移：User → Post → Comment（外键依赖父表）
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{})
	if err != nil {
		log.Fatalf("表结构迁移失败：%v", err)
	}
	log.Println("表结构创建/更新成功！")

	return db
}

// -------------------------- 插入数据 --------------------------
func Insert(db *gorm.DB) error {
	user := User{
		Username: "root",
		Posts: []Post{
			{Title: "爱国1", Content: "我很爱国1", Comments: []Comment{
				{Content: "我很爱国1的评论1", UserID: 1},
				{Content: "我很爱国1的评论2", UserID: 1},
				{Content: "我很爱国1的评论3", UserID: 1},
			}},
			{Title: "爱国2", Content: "我很爱国2", Comments: []Comment{
				{Content: "我很爱国2的评论1", UserID: 1},
				{Content: "我很爱国2的评论2", UserID: 1},
			}},
		},
	}
	db.Create(&user)
	log.Println("表数据创建成功！")
	return nil
}

// -------------------------- 需求1：查询某个用户的所有文章及对应评论 --------------------------
// GetUserPostsWithComments 根据用户ID查询该用户的所有文章，且每篇文章包含对应的评论列表
func GetUserPostsWithComments(db *gorm.DB, userID uint) ([]Post, error) {
	var user User
	// 1. Preload 嵌套预加载：先加载用户的 Posts，再加载每个 Post 的 Comments
	// 等价于：查询用户 -> 批量查询该用户的所有文章 -> 批量查询这些文章的所有评论
	if err := db.Preload("Posts.Comments").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("查询用户失败：%v", err)
	}
	return user.Posts, nil
}

// -------------------------- 需求2：查询评论数量最多的文章信息 --------------------------
// GetPostWithMostComments 查询评论数最多的文章（含评论数、文章基本信息）
// 返回：文章信息、评论总数、错误
func GetPostWithMostComments(db *gorm.DB) (Post, int64, error) {
	var post Post
	var commentCount int64

	// 1. 子查询：统计每篇文章的评论数，按评论数降序，取第一条
	subQuery := db.Model(&Comment{}).Select("post_id, COUNT(*) as comment_count").Group("post_id")

	// 2. 关联查询：将文章表与子查询结果关联，按评论数降序，取第一条
	err := db.Model(&Post{}).
		Joins("LEFT JOIN (?) as comment_stats ON posts.id = comment_stats.post_id", subQuery).
		Select("posts.*, IFNULL(comment_stats.comment_count, 0) as comment_count").
		Order("comment_count DESC").
		First(&post).
		// 提取评论数（需给 Post 临时加 comment_count 字段，或单独查询）
		// 此处通过 Raw 更直观，或用 Scan 到自定义结构体
		Error

	if err != nil {
		// 处理无数据的情况
		if err == gorm.ErrRecordNotFound {
			return post, 0, fmt.Errorf("暂无文章数据")
		}
		return post, 0, fmt.Errorf("查询评论最多的文章失败：%v", err)
	}

	// 单独查询该文章的评论总数（更稳妥的方式）
	db.Model(&Comment{}).Where("post_id = ?", post.ID).Count(&commentCount)

	return post, commentCount, nil
}

func DeleteComments(db *gorm.DB, PostID uint) {
	if PostID == 0 {
		return
	}
	//db.Unscoped().Model(&Comment{}).Where("post_id = ?", PostID).UpdateColumn("deleted_at", nil)
	comment := Comment{PostID: PostID}
	db.Where("post_id = ?", PostID).Delete(&comment)
	log.Println("删除成功")

}
