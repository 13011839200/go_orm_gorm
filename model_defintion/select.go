package model_defintion

import (
	"gorm.io/gorm"
)

func Select(db *gorm.DB) {
	//userName := "张三"
	//user := User{Name: userName}
	//db.Find(&user)
	//fmt.Println("用户", userName, "的信息为:", user)
	//
	//posts := make([]Post, 0)
	//db.Where("user_id = ?", user.ID).Find(&posts)
	//
	//for _, post := range posts {
	//	fmt.Println("用户", userName, "的文章内容为：", post.Content)
	//	comments := make([]Comment, 0)
	//
	//	db.Where("post_id = ?", post.ID).Find(&comments)
	//	for _, comment := range comments {
	//
	//		fmt.Println("文章", post.Content, "的评论为:", comment.CommentContent)
	//	}
	//}

}
