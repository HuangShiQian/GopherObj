package controllers

import (
	"github.com/astaxie/beego"
	"path"
	"time"
	"github.com/astaxie/beego/orm"
	"code2/newsWeb/models"
	"math"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"bytes"
	"encoding/gob"
)

type ArticleController struct {

	beego.Controller
}

//展示首页的方法
func (this *ArticleController) ShowIndex() {
	//校验登录状态
	userName:=this.GetSession("userName")//返回的userName为接口类型
	if userName==nil{
		this.Redirect("/login",302)
		return
	}
	this.Data["userName"]=userName.(string)//断言

	//获取所有文章数据  展示到页面
	o:=orm.NewOrm()
	qs:=o.QueryTable("Article")
	//定义切片
	var articles  []models.Article
	//qs.All(&articles)
	//this.Data["articles"]=articles//把数据传给前端
	//this.TplName = "index.html"
	//获取总记录数
	//count,_ := qs.RelatedSel("ArticleType").Count()  //这行代码在获取选中类型后不可用 有bug 要换成下方0506的代码
	//获取总页数
	//获取选中的类型  0506
	typeName:=this.GetString("select") //获取选中类型
	var count int64
	if typeName == "" {
		//获取总记录数
		count,_ = qs.RelatedSel("ArticleType").Count()
	}else {
		count,_ = qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).Count()
	}//0506

	pageIndex := 2


	pageCount := math.Ceil(float64(count) / float64(pageIndex))
	//获取首页和末页数据
	//获取页码
	pageNum ,err := this.GetInt("pageNum")
	if err != nil {
		pageNum = 1
	}
	beego.Info("数据总页数为:",pageNum)

	//获取对应页的数据   获取几条数据     起始位置
	//ORM多表查询的时候默认是惰性查询  关联查询之后  如果关联的字段为空 数据查询不到



	//where ArticleType.typeName = typename   filter相当于where
	if typeName==""{
		qs.Limit(pageIndex,pageIndex * (pageNum - 1)).RelatedSel("ArticleType").All(&articles)
	}else {
		qs.Limit(pageIndex,pageIndex * (pageNum - 1)).RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).All(&articles)
	}



	//查询所有文件类型 并展示
	var articleTypes  []models.ArticleType //结构体切片
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"]=articleTypes

	//把数据集存到redis中
	conn,err:=redis.Dial("tcp",":6379")
	if err!=nil{
		beego.Error("redis连接失败")
		return
	}
	defer conn.Close()

	resp,err:=conn.Do("get","newsWeb")//获取数据
	result,_:=redis.Bytes(resp,err)//回复助手函数
	if len(result)==0{//没有取到数据  说明是第一次访问  从MySQL中拿数据
		o.QueryTable("ArticleType").All(&articleTypes)
		this.Data["articleTypes"] = articleTypes
		//把数据编码后存到redis中
		var buffer bytes.Buffer//把数据存到Buffer结构体
		enc:=gob.NewEncoder(&buffer)
		enc.Encode(articleTypes)//把articleTypes转化成二进制数据
		conn.Do("set","newsWeb",buffer.Bytes())//buffer.Bytes()表示把buffer结构体转成二进制数据  并存到redis
		beego.Info("从MySQL中获取数据")
	}else {//取到了数据 说明不是第一次访问 从Redis中拿数据
		//解码  把二进制类型转化成最终需要类型
		dec:=gob.NewDecoder(bytes.NewReader(result))
		dec.Decode(&articleTypes)
		beego.Info("从Redis中获取数据")
	}

	/*//conn.Do("set","newsWeb",articleTypes)

    //把数据存储到redis中
	//序列化
	//1.先定义容器  用来接收编码后的二进制数据
	var buffer bytes.Buffer
	//2.获取编码器
	enc:=gob.NewEncoder(&buffer)
	//3.编码
	enc.Encode(articleTypes)//articleTypes为要编码的数据
	conn.Do("set","newsWeb",buffer.Bytes())

	//反序列化 解码操作
	//先拿数据
	resp,err:=conn.Do("get","newsWeb")
	//回复助手
	result,_:=redis.Bytes(resp,err)//转成字节
	//解码
	//1.定义一个解码器
	dec:=gob.NewDecoder(bytes.NewReader(result))
	//定义一个容器 接收编码的数据
	var newsTypes []models.ArticleType
	//2.进行解码操作
	dec.Decode(&newsTypes)
	beego.Info(newsTypes)//从redis中拿数据*/


	this.Data["articles"] = articles  //把数据传给前端
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount
	this.Data["pageNum"] = pageNum//给前端传页码

	this.Data["TypeName"] = typeName//0506加上  给前端文章分类option处传typeName
	this.Layout="layout.html"
	this.TplName = "index.html"

}

//按照类型展示首页  (这个是否可以直接删掉)
func(this*ArticleController)HandleIndex(){

}

//展示 添加文章页面 业务
func (this *ArticleController) ShowAddArticle() {
	//获取所有类型并绑定下拉框
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)

	this.Data["articleTypes"] = articleTypes
	this.Layout="layout.html"
	this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandleAddArticle() {
	//	1.获取表数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	typeName := this.GetString("select")//0505

	//	2.校验数据
	if articleName == "" || content == "" || typeName == ""{// || typeName == ""    0505日写
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "获取数据错误" //传数据给页面
		this.TplName = "add.html"
		return
	}
	//	3.获取图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片上传失败" //传数据给页面
		this.TplName = "add.html"
		return
	}
	defer file.Close() //关闭文件

	//保存图片之前 判断文件各种数据情况

	//第一 判断文件数据大小
	if head.Size > 5000000 {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片数据过大" //传数据给页面
		this.TplName = "add.html"
		return
	}

	//第二 校验数据格式
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "上传图片格式错误" //传数据给页面
		this.TplName = "add.html"
		return
	}
	//第三 防止重名
	fileName := time.Now().Format("2006-01-02-15-04-05-2222")
	//这一行的资源路径前面的.不要忘记了 beego的bug
	//资源路径最后的/ 也千万不要忘记   (如果忘记了 就跑到同级目录去了)
	err = this.SaveToFile("uploadname", "./static/img/"+fileName+ext)
	if err != nil {
		beego.Error("保存图片失败")
		this.Data["errmsg"] = "保存图片失败" //传数据给页面
		this.TplName = "add.html"
		return
	}

	//	4.处理数据
	//  01-获取orm对象
	o := orm.NewOrm()
	//02-获取插入对象
	var article models.Article
	//	03-给插入对象赋值
	article.Title = articleName
	article.Content = content
	article.Img = "/static/img/" + fileName + ext

	//获取一个类型对象，并插入到文章中  0505日下午
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Read(&articleType,"TypeName")

	article.ArticleType = &articleType

	// 04-插入数据
	_, err = o.Insert(&article)
	if err != nil {
		beego.Error("保存图片失败")
		this.Data["errmsg"] = "数据插入失败" //传数据给页面
		this.TplName = "add.html"
		return
	}



	//5.返回数据
	this.Redirect("/article/index", 302)
}

//查看文章详情
func (this*ArticleController) ShowContent()  {
//	获取数据
id,err:=this.GetInt("id")
	//校验数据
	if err!=nil{
		beego.Error("获取文章id错误")
		this.Redirect("/article/index",302)//渲染  如果页面本身有数据加载，不能直接渲染
		return
	}

	//处理数据
	//查询文章数据
	o:=orm.NewOrm()
	//获取查询对象
	var article  models.Article
	//给查询条件赋值
	article.Id=id
	//查询
	o.Read(&article)
	//多对多查询一
	//o.LoadRelated(&article,"Users")
	//高级查询 首先要指定表  多对多查询二  获取用户名    为了使用高级查询
	var users []models.User
	o.QueryTable("User").Filter("Articles__Article__Id",id).Distinct().All(&users)
	this.Data["users"]=users


	//给更新条件赋值
	article.ReadCount+=1
	o.Update(&article)

	//返回数据
	this.Data["article"]=article

	//插入多对多关系  根据用户名获取用户对象
	userName:=this.GetSession("userName")
	var user  models.User
	user.Name=userName.(string)//类型断言
	o.Read(&user,"Name")//非主键时要指定字段

	//多对多的插入操作 分三步
	//1.获取ORM对象

	//2.获取被插入数据的对象  文章

	//3.获取多对多操作对象
	m2m:=o.QueryM2M(&article,"Users")
	//用多对多操作对象插入
	m2m.Add(user)

	this.Layout="layout.html"
	this.TplName="content.html"
}

//展示文章编辑页
func (this*ArticleController)ShowUpdate()  {
	//获取数据
	id,err:=this.GetInt("id")
	//校验数据
	if err!=nil {
		beego.Error("获取文章id错误")
		this.Redirect("/article/index",302)
		return
	}

//	处理数据
o:=orm.NewOrm()
	var article models.Article
	article.Id=id
	o.Read(&article)
	//返回数据
	this.Data["article"] = article
	this.TplName = "update.html"
}

//封装上传文件处理函数
func UploadFile(this *ArticleController,filePath string,errHtml string)string{
	//获取图片
	//返回值 文件二进制流  文件头    错误信息
	file,head,err := this.GetFile(filePath)
	if err != nil {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片上传失败"
		this.TplName = errHtml
		return ""
	}
	defer file.Close()
	//校验文件大小
	if head.Size >5000000{
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片数据过大"
		this.TplName = errHtml
		return ""
	}

	//校验格式 获取文件后缀
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "上传文件格式错误"
		this.TplName = errHtml
		return ""
	}

	//防止重名
	fileName := time.Now().Format("200601021504052222")


	//jianhuangcaozuo

	//把上传的文件存储到项目文件夹
	this.SaveToFile(filePath,"./static/img/"+fileName+ext)
	return "/static/img/"+fileName+ext

}

//处理文章编辑
func (this*ArticleController) HandleUpdate() {
	//获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	savePath := UploadFile(this,"uploadname","update.html")
	id,_ := this.GetInt("id")  //隐藏域传值
	//校验数据
	if articleName == "" || content == "" ||savePath == "" {
		beego.Error("获取数据失败")
		this.Redirect("/article/update?id="+strconv.Itoa(id),302)
		return
	}
	//处理数据
	//更新操作
	o := orm.NewOrm()
	var article models.Article
	//先查询要更新的文章是否存在
	article.Id = id
	//必须查询
	o.Read(&article)
	//更新   需要先赋新值   beego中的ORM如果需要更新，更新的对象Id必须有值
	article.Title = articleName
	article.Content = content
	article.Img = savePath
	o.Update(&article)


	//返回数据
	this.Redirect("/article/index",302)
}

//删除文章
func(this*ArticleController)HandleDelete(){
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取Id错误")
		this.Redirect("/article/index",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Delete(&article,"Id")

	//返回数据
	this.Redirect("/article/index",302)
}

//展示添加分类页面
func (this*ArticleController)ShowAddType()  {
	//获取所有类型，并展示到页面上
	//获取所有用all
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)

	//返回数据
	this.Layout="layout.html"
	this.Data["articleTypes"] = articleTypes

	this.TplName = "addType.html"
}
//处理添加类型请求
func (this*ArticleController)HandleAddType()  {
	//获取数据
	typeName := this.GetString("typeName")
	//校验数据
	if typeName == ""{
		beego.Error("类型名称传输失败")
		this.Redirect("/article/addType",302)
		return
	}
	//处理数据
	//插入操作
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Insert(&articleType)

	//返回数据
	this.Redirect("/article/addType",302)
}

//删除类型
func (this*ArticleController)DeleteType()  {
//	获取数据
id,err:=this.GetInt("id")
//校验数据
if err!=nil{
	beego.Error("获取文章id失败")
	this.Redirect("/article/addType",302)
	return
}
//处理数据  删除数据
o:=orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id=id
	o.Delete(&articleType,"Id")
//返回数据
this.Redirect("/article/addType",302)

}