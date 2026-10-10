package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//题目1：模型定义
//假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
//要求 ：
//使用Gorm定义 User 、 Post 和 Comment 模型，其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章）， Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
//编写Go代码，使用Gorm创建这些模型对应的数据库表。
//题目2：关联查询
//基于上述博客系统的模型定义。
//要求 ：
//编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
//编写Go代码，使用Gorm查询评论数量最多的文章信息。
//题目3：钩子函数
//继续使用博客系统的模型。
//要求 ：
//为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
//为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"。

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Posts     []Post `gorm:"foreignkey:UserID"`
	PostCount int    `gorm:"default 0"`
}

type Post struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"size:255;not null;index"`
	Content       string    `gorm:"type:text"`
	UserID        uint      `gorm:"index;not null"`
	CommentStatus string    `gorm:"size:255;default:'无评论'"`
	Comments      []Comment `gorm:"foreignkey:PostID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text;not null"`
	PostID    uint   `gorm:"index;not null"`
	UserID    uint   `gorm:"index;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func SeedData(db *gorm.DB) {
	// ---------- 0. 已经有数据就跳过（防止重复运行造成数据混乱） ----------
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount > 0 {
		fmt.Println("ℹ️  数据库已有数据，跳过填充")
		return
	}
	// ---------- 1. 插入用户 ----------
	users := []User{
		{Name: "张三", Email: "zhangsan@example.com"},
		{Name: "李四", Email: "lisi@example.com"},
		{Name: "王五", Email: "wangwu@example.com"},
	}
	if err := db.Create(&users).Error; err != nil {
		fmt.Println("❌ 插入用户失败:", err)
		return
	}
	fmt.Printf("✅ 插入 %d 个用户\n", len(users))
	// Create(&users) 之后，GORM 会把自增 ID 回填到切片里，
	// 所以下面可以直接用 users[0].ID、users[1].ID 等
	// ---------- 2. 插入文章 ----------
	posts := []Post{
		{Title: "Go 语言入门", Content: "Go 是一门由 Google 开发的静态类型语言...", UserID: users[0].ID},
		{Title: "GORM 实战", Content: "GORM 是 Go 最流行的 ORM 框架...", UserID: users[0].ID},
		{Title: "MySQL 索引优化", Content: "索引可以极大提升查询效率...", UserID: users[0].ID},
		{Title: "Redis 缓存设计", Content: "Redis 是一个内存数据库...", UserID: users[1].ID},
	}
	if err := db.Create(&posts).Error; err != nil {
		fmt.Println("❌ 插入文章失败:", err)
		return
	}
	fmt.Printf("✅ 插入 %d 篇文章\n", len(posts))
	// ---------- 3. 插入评论 ----------
	// 故意让「文章1」评论最多，方便后面测试「评论数最多的文章」
	// 文章4 故意一条评论都没有，方便测试「无评论」状态
	comments := []Comment{
		// 文章1：4 条评论（最多）
		{Content: "写得真好，学到了！", PostID: posts[0].ID, UserID: users[1].ID},
		{Content: "请问有配套代码吗？", PostID: posts[0].ID, UserID: users[2].ID},
		{Content: "收藏了，感谢分享", PostID: posts[0].ID, UserID: users[1].ID},
		{Content: "期待后续！", PostID: posts[0].ID, UserID: users[0].ID},
		// 文章2：2 条评论
		{Content: "GORM 确实好用", PostID: posts[1].ID, UserID: users[1].ID},
		{Content: "Preload 那部分讲得很清楚", PostID: posts[1].ID, UserID: users[2].ID},
		// 文章3：1 条评论
		{Content: "索引原理能再深挖一点吗？", PostID: posts[2].ID, UserID: users[2].ID},
		// 文章4：0 条评论（故意留空）
	}
	if err := db.Create(&comments).Error; err != nil {
		fmt.Println("❌ 插入评论失败:", err)
		return
	}
	fmt.Printf("✅ 插入 %d 条评论\n", len(comments))
	// ---------- 4. 同步统计字段 ----------
	// 如果你还没做「题目3」的钩子，PostCount / CommentStatus 可能还是默认值。
	// 这里统一刷一遍，保证数据和真实情况一致。
	syncStats(db)
	fmt.Println("🎉 测试数据填充完成！")
}

// syncStats 根据真实数据，重新计算每个用户的 post_count、
// 每篇文章的 comment_status
func syncStats(db *gorm.DB) {
	// ---- 同步每个用户的文章数 ----
	var users []User
	db.Find(&users)
	for _, u := range users {
		var cnt int64
		db.Model(&Post{}).Where("user_id = ?", u.ID).Count(&cnt)
		db.Model(&User{}).Where("id = ?", u.ID).
			UpdateColumn("post_count", cnt)
	}
	// ---- 同步每篇文章的评论状态 ----
	var posts []Post
	db.Find(&posts)
	for _, p := range posts {
		var cnt int64
		db.Model(&Comment{}).Where("post_id = ?", p.ID).Count(&cnt)
		status := "无评论"
		if cnt > 0 {
			status = "有评论"
		}
		db.Model(&Post{}).Where("id = ?", p.ID).
			UpdateColumn("comment_status", status)
	}
	fmt.Println("✅ 统计字段已同步")
}

func GetUserPostsandComments(db *gorm.DB, userID uint) error {
	var user User
	err := db.Preload("Posts",
		func(db *gorm.DB) *gorm.DB {
			return db.Order("posts.created_at DESC")
		}).Preload("Posts.Comments",
		func(db *gorm.DB) *gorm.DB {
			return db.Order("comments.created_at DESC")
		}).First(&user, userID).Error
	if err != nil {
		log.Fatalf("search data failed: %v", err)
		return err
	}
	fmt.Printf("用户【%s】一共有%d篇文章: \n", user.Name, len(user.Posts))
	for _, p := range user.Posts {
		fmt.Printf("--文章【%s】一共有评论%d条: \n", p.Title, len(p.Comments))
		for _, c := range p.Comments {
			fmt.Printf("--评论: %s \n", c.Content)
		}
		fmt.Println("=========================\n")
	}

	return nil
}

func GetPostsWithMostComments(db *gorm.DB) (*Post, error) {
	type CountResult struct {
		PostID uint
		Total  int
	}

	var count CountResult
	err := db.Model(&Comment{}).
		Select("post_id,count(*) as total").
		Group("post_id").
		Order("total DESC").
		Limit(1).Scan(&count).Error
	if err != nil {
		log.Fatalf("search data failed: %v", err)
		return nil, err
	}
	if count.Total == 0 {
		return nil, errors.New("暂时没有任何文章有评论")
	} else {
		var post Post
		if err := db.Preload("Comments").First(&post, count.PostID).Error; err != nil {
			log.Fatalf("search data failed: %v", err)
			return nil, err
		}
		return &post, nil
	}
}

func (p *Post) AfterCreate(tx *gorm.DB) error {
	if p.UserID == 0 {
		return nil
	}
	err := tx.Model(&User{}).Where("id = ?", p.UserID).Update("post_count", gorm.Expr("post_count + 1")).Error
	if err != nil {
		log.Fatalf("update post count failed: %v", err)
		return err
	}
	return nil
}

func (c *Comment) AfterDelete(tx *gorm.DB) error {
	if c.PostID == 0 {
		return nil
	}

	var count int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count).Error; err != nil {
		log.Fatalf("search data failed: %v", err)
		return err
	}
	status := "有评论"
	if count == 0 {
		status = "无评论"
	}

	if err := tx.Model(&Post{}).Where("id = ?", c.PostID).UpdateColumn("comment_status", status).Error; err != nil {
		log.Fatalf("update comment status failed: %v", err)
		return err
	}
	return nil
}

func main() {
	db, err := gorm.Open(sqlite.Open("homework03.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{}, &Post{}, &Comment{})
	fmt.Println("Database created")

	SeedData(db)
	fmt.Println("Seed done")

	GetUserPostsandComments(db, 1)

	var post *Post
	if post, err = GetPostsWithMostComments(db); err != nil {
		log.Fatalf("search data failed: %v", err)
	} else {
		fmt.Printf("评论最多的文章id为%d", post.ID)
	}

}
