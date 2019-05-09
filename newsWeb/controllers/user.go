package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"code2/newsWeb/models"
	"encoding/base64"
)

type UserController struct {//注意 现有类结构体 才能创建这个控制器下面的方法
	beego.Controller
}
func (this *UserController)ShowRegister()  {
	//展示注册页面
	this.TplName="register.html"
}

//处理注册业务
func (this*UserController)HandleRegister()  {
//	1.获取数据
	userName:=this.GetString("userName")
	pwd:=this.GetString("password")
//2.校验数据
if userName==""||pwd==""{
	beego.Error("传输数据不完整")
	this.TplName="register.html"
	return
}
//3.处理数据
   o:=orm.NewOrm()
	var user models.User
	user.Name=userName
	user.Pwd=pwd
	id,err:=o.Insert(&user)
	if err!=nil{
		beego.Error("用户注册失败")
		this.TplName="register.html"
		return
	}
	beego.Info(id)
//	返回数据
//this.TplName="login.html"
//跳转 函数 参数1：新指向的地址"/login"   参数2： 状态码 302 表重定向
this.Redirect("/login",302)
}


//展示登录页面
func (this*UserController)ShowLogin()  {
	//获取cookie数据，如果获取查到了，说明上一次记住用户名，不然的话，不记住用户名
	userName:=this.Ctx.GetCookie("userName")
	//解密
	dec,_:=base64.StdEncoding.DecodeString(userName)
	if userName!=""{
		this.Data["userName"]=string(dec)
		this.Data["checked"]="checked"
	}else {
		this.Data["userName"]=""
		this.Data["checked"]=""
	}
	this.TplName="login.html"
}

//处理登录业务
func (this*UserController)HandleLogin()  {
	//1.获取数据
	userName:=this.GetString("userName")
	pwd:=this.GetString("password")
//	2.校验数据
if userName==""||pwd==""{
	beego.Error("传输数据不完整")
	this.TplName="login.html"
	return
}
//3.处理数据
o:=orm.NewOrm()
	var  user models.User
	user.Name=userName
	err:=o.Read(&user,"Name")
	if err!=nil{
		beego.Error("用户名不存在")
		this.TplName="login.html"
		return
	}

	if user.Pwd!=pwd{
		beego.Error("密码错误")
		this.TplName="login.html"
		return
	}
	//实现记住用户名功能  上一次登陆成功以后，点击了记住用户名，下一次登陆的时候默认显示用户名0506

	remember:=this.GetString("remember")
	//base64:把一些非常见字符转成常见字符
	//给userName加密
	enc:=base64.StdEncoding.EncodeToString([]byte(userName))
	if remember=="on"{
		this.Ctx.SetCookie("userName",enc,60)
	}else {
		this.Ctx.SetCookie("userName",userName,-1)//-1表示使cookie失效 即删除cookie
	}

	//Session存储
	this.SetSession("userName",userName)
//4.返回数据
//this.Ctx.WriteString("用户登录成功")
this.Redirect("/article/index",302)
}


//退出登录业务
func (this*UserController)Logout()  {
	//删除session 然后跳转到登录页面
	this.DelSession("userName")
	this.Redirect("/article/login",302)
}