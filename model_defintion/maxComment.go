package model_defintion

import (
	"fmt"

	"gorm.io/gorm"
)

func MaxComment(db *gorm.DB) {
	var postId uint
	db.Raw("SELECT post_id AS postId FROM comments  GROUP BY post_id  HAVING COUNT(id) = (SELECT MAX(countId) FROM (SELECT COUNT(id) AS countId FROM comments GROUP BY post_id) AS c)").Find(&postId)
	post := Post{}
	db.Find(&post, postId)
	fmt.Println("评论数量最多的文章内容为：", post.Content)
}
