package main_config

import (
	"time"
	vllog "common/villog"
	"flag"
	"path"
	"os"
	"io"
	"encoding/base64"
	"github.com/shadowsocks/go-shadowsocks2/core"
	"crypto/rand"
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

var MConfig = &Config{}


var InputArgs struct {
	HomePath string
	ConfPath string
}

func LoadConfigFile(configName string){
	if err := vllog.LoadJsonData(configName, MConfig); err != nil{
		vllog.LogE("Load config file error:%v", err)
		panic(0)
	}
}


/**
 * 参数和配置读取
 */
func ArgsPare() *Config{

	//设置命令行参数
	flag.StringVar(&InputArgs.HomePath, "home", "", "The home path of progream")
	flag.StringVar(&InputArgs.ConfPath, "conf", "", "The config file path")
	flag.DurationVar(&MConfig.UDPTimeout,"timeout", 5*time.Minute, "UDP tunnel timeout")
	flag.Parse()

	vllog.LogI("InputArgs.HomePath:%s", InputArgs.HomePath)
	vllog.LogI("InputArgs.ConfPath:%s", InputArgs.ConfPath)
	vllog.LogI("config.UDPTimeout:%s", MConfig.UDPTimeout)

	if(InputArgs.ConfPath == ""){
		LoadConfigFile(path.Join(InputArgs.HomePath ,"./config/config.json"))
	}else{
		LoadConfigFile(InputArgs.ConfPath)
	}

	for k, user := range MConfig.UserGroups{
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

	return MConfig
}