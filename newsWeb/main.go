package main

import (
	_ "code2/newsWeb/routers"
	"github.com/astaxie/beego"
	_"code2/newsWeb/models"
	_"code2/newsWeb/controllers"
)

func main() {

	beego.AddFuncMap("prePage",PrePage)
	beego.AddFuncMap("nextPage",NextPage)
	beego.Run()
}
//以下两个函数为什么要写在main.go里
func PrePage(pageNum int)int  {
	if pageNum<=1{
		return 1
	}
	return pageNum-1
}

func NextPage(pageNum int,pageCount float64) int {
	if pageNum>=int(pageCount) {
		return int(pageCount)
	}
	return pageNum+1
}