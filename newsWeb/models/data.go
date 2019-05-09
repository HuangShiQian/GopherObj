package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
//用户表
type User struct {
	Id   int
	Name string
	Pwd  string
	Articles []*Article `orm:"reverse(many)"`
}

func init() {  //初始化数据库三行代码
	//连接数据库   注意这里的数据库默认端口3306 即要写：root:123456@tcp(127.0.0.1:3306)/newsWeb
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/newsWeb")
	//	注册表
	orm.RegisterModel(new(User),new(Article),new(ArticleType))
	//run
	orm.RunSyncdb("default", false, true)

}

//创建添加文章表
type Article struct {
	Id int `orm:"pk;auto"`
	Title string `orm:"unique;size(40)"`
	Content string `orm:"size(500)"`
	Img string `orm:"null"`//表示允许为空
	Time time.Time `orm:"type(datetime);auto_now_add"`
	ReadCount int `orm:"default(0)"`
//	这里结束后 回到上面orm.RegisterModel(new(User),new(Article)) 注册表处加上new(Article)
	ArticleType *ArticleType `orm:"rel(fk);null;on_delete(set_null)"`
	Users []*User `orm:"rel(m2m)"`

}


//创建类型表
type ArticleType struct {
	Id int
	TypeName string
	Article []*Article `orm:"reverse(many)"`
}
