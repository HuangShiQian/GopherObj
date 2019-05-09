package controllers

import (
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego"
)

func init()  {
	//连接函数
	conn,err:=redis.Dial("tcp",":6379")
	if err!=nil{
		beego.Error("redis连接失败")
	}
	//操作函数
	//resp,err:=conn.Do("mget","t1","t3")
	//result,_:=redis.Strings(resp,err)
	//beego.Info("获取的数据为",result)
//t1 11   t2 bj3q   t3 90   这里t1和t3同是int类型  t3为string类型
	resp,err:=conn.Do("mget","t1","t2","t3")
	//回复助手函数
	result,_:=redis.Values(resp,err)
	//把对应的函数扫描到相应变量里面
	var v1 ,v2 int
	var v3 string
	redis.Scan(result,&v1,&v3,&v2)

	beego.Info("获取的数据为",v1,v2,v3)



}
