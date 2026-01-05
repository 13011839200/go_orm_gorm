package main

import (
	"go_orm_gorm/model_defintion"
)

func main() {
	// 初始化数据库
	db := model_defintion.InitDB()
	// 插入数据
	//_ = model_defintion.Insert(db)

	// 查询某个用户的所有文章及对应评论
	//posts, err := model_defintion.GetUserPostsWithComments(db, 1)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(posts)
	//
	//// 查询评论数量最多的文章信息
	//post, count, err := model_defintion.GetPostWithMostComments(db)
	//if err != nil {
	//	return
	//}
	//fmt.Println(post)
	//fmt.Println(count)
	model_defintion.DeleteComments(db, 1)

}
