package main

import (
	"path"
	"time"
	"github.com/shadowsocks/go-shadowsocks2/core"
	vllog "villeSock/vender/common/villog"
	"flag"
	//"strings"
	"net/url"
	"io"
	"encoding/base64"
	"crypto/rand"
	"os"
	"syscall"
	"fmt"
	"os/signal"
	"strings"
)

/**
 * 用户组
 */
type UserGroup struct{
	Name string `json:"name"`
	Server    string `json:"server"`
	Port int	`json:"port"`
	Password  string `json:"password"`
	Cipher    string `json:"cipher"`
	Key       string `json:"key"`
	Keygen    int	 `json:"key_gen"`
	UDPTimeout	  time.Duration	 `json:"time_out"`
}

type Config struct {
	UserGroups []*UserGroup `json:"user_groups"`
	UDPTimeout time.Duration `json:"time_out"`
}

var config = &Config{}


var InputArgs struct {
	HomePath string
	ConfPath string
}

func LoadConfigFile(configName string){
	//
	if err := vllog.LoadJsonData(configName, config); err != nil{
		vllog.LogE("Load config file error:%v", err)
		panic(0)
	}
}

/**
 * 参数和配置读取
 */
func argsPare(){
	//设置命令行参数
	flag.StringVar(&InputArgs.HomePath, "home", "", "The home path of progream")
	flag.StringVar(&InputArgs.ConfPath, "conf", "", "The config file path")
	flag.DurationVar(&config.UDPTimeout,"timeout", 5*time.Minute, "UDP tunnel timeout")
	flag.Parse()

	vllog.LogI("InputArgs.HomePath:%s", InputArgs.HomePath)
	vllog.LogI("InputArgs.ConfPath:%s", InputArgs.ConfPath)
	vllog.LogI("config.UDPTimeout:%s", config.UDPTimeout)

	if(InputArgs.ConfPath == ""){
		LoadConfigFile(path.Join(InputArgs.HomePath ,"./config/config.json"))
	}else{
		LoadConfigFile(InputArgs.ConfPath)
	}


	for k, user := range config.UserGroups{

		if user.Server == "" || user.Server == " "{
			vllog.LogE("config.Server is empty:%s")
			os.Exit(1)
		}

		if user.Keygen > 0 {
			key := make([]byte, user.Keygen)
			io.ReadFull(rand.Reader, key)
			user.Key = base64.URLEncoding.EncodeToString(key)
		}
		if user.Cipher == "" || user.Cipher == " " {
			vllog.LogE("cipher must be %v", core.ListCipher())
			os.Exit(1)
		}
		vllog.LogI("\n[No.%d]\n[Name:%s]\n[Server:%s]\n[Port:%d]\n[Password:%s]\n[Cipher:%s]\n[Key:%s]",k,
			user.Name, user.Server, user.Port, user.Password, user.Cipher, string(user.Key))
	}

}


func main() {
	//获取参数
	argsPare();



	/**
	 * ciph ：Cipher interface类型 主要包含 StreamConn(net.Conn) net.Conn
	 * 和 PacketConn(net.PacketConn) net.PacketConn
	 */
	 for _, user := range config.UserGroups{

		 addr := user.Server
		 cipher := user.Cipher
		 password := user.Password
		 var err error
		 //判断是否以 ss://开头 (ss://是 websocket连接协议)
		 if strings.HasPrefix(addr, "ss://") {
			 addr, cipher, password, err = parseURL(addr)
			 if err != nil {
				 vllog.LogE("error :",err)
			 }
		 }else{
			 addr = fmt.Sprintf("%s:%d",user.Server, user.Port)
		 }
		 ciph, err := core.PickCipher(cipher, []byte(user.Key), password)
		 if err != nil {
			 vllog.LogE("Error :", err)
			 os.Exit(1)
		 }
		 vllog.LogI("addr = %s ciph=%v",addr, ciph)
		 go udpRemote(addr, user.UDPTimeout ,ciph.PacketConn)
		 go tcpRemote(addr, ciph.StreamConn)
	 }

	//检测系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	processStr := <-sigCh
	fmt.Printf("退出 shadowsocket-go 进程 %v", processStr );
}

/**
 * @author：注释-凌霄
 * 解析 url 参数 参数格式为：scheme://[userinfo@]host/path[?query][#fragment]
 * @return addr: host+post ,
 * @return cipher: 用户名
 * @return password:密码
 * @return err 错误信息
 */
func parseURL(s string) (addr, cipher, password string, err error) {
	/**
		URL类型代表一个解析后的URL（或者说，一个URL参照）。
		URL基本格式如下：
		scheme://[userinfo@]host/path[?query][#fragment]
		例如："postgres://user:pass@host.com:5432/path?k=v#f"
		Host 同时包括主机名和端口信息，如过端口存在的话，使用 strings.Split() 从 Host 中手动提取端口
		User 包含了所有的认证信息，这里调用 Username和 Password 来获取独立值
	 */
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}
