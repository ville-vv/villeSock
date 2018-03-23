## villeSock 

根据 go-shadowsocket2 修改的代码，添加了日志和多端口支持。之前是通过启动service的时候传入参数来配置端口的，现在是通过配置json文件来支持的。

配置文件配置示例：

```json
{
    "time_out":20,
    "user_groups":[
        {
            "server":"0.0.0.0",
            "name":"myown",
            "port":8080,
            "password":"1234567",
            "cipher":"AEAD_CHACHA20_POLY1305",
            "time_out":10
        }
    ]
}
```

读取配置文件的位置可以通过启动参数自定义。

```
vilSock -conf "路径+文件名"
```



