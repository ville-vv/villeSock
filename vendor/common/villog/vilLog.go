package villog

import (
	"github.com/op/go-logging"
	"io/ioutil"
	"encoding/json"
	"os"
	"net"
	"strings"
	"fmt"
	"path"
	"io"
)

/**
 *日志级别
 */
const (
	CRITICAL int = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

/**
 * 日志输出格式
 */
var LogFormat = []string{
	`%{shortfunc} ▶ %{level:.4s} %{message}`,
	`%{time:15:04:05.00} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}`,
	`%{color}%{time:15:04:05.00} %{shortfunc} %{shortfile} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
}

/**
 * 日志级别与 string类型映射
 */
var LogLevelMap = map[string]int{
	"CRITICAL":CRITICAL,
	"ERROR":ERROR,
	"WARNING":WARNING,
	"NOTICE":NOTICE,
	"INFO":INFO,
	"DEBUG":DEBUG,
}

/**
 * 从文件读取json 数据
 * @fileName 	:	文件名称
 * @v 			：	保存json文件的数据结构
 */
func LoadJsonData(fileName string, v interface{}) error{
	data, err := ioutil.ReadFile(fileName)
	if err != nil{
		return err;
	}

	dataJson := []byte(data)

	if err = json.Unmarshal(dataJson, v); err != nil{
		return err
	}
	return nil
}

/**
 *获取本机IP
 */
func getServerIP() (ip string, err error) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		return
	}
	defer conn.Close()
	ip = strings.Split(conn.LocalAddr().String(), ":")[0]

	return
}

/**
 * 获取本机名字
 */
func getHostName() (hostname string, err error) {
	hostname, err = os.Hostname()
	return
}

/**
 * 日志配置文件结构体
 */
type VilLogConf struct{
	ConfFilePath string `json:"conf_file_path"`		//配置文件路径
	Path      string `json:"out_path"`				//日志输出路径
	Level     string `json:"log_level"`
	FormatMd    int `json:"format_md"`
	FileBackEnd bool `json:"file_back_end"`			//文件终端输出
	StderrBackEnd bool `json:"stderr_back_end"`		//标准错误终端输出
	NetBackEnd bool `json:"net_back_end"`			// 网络终端输出
	ModuleName string `json:"module_name"`			// 模块名称
	ExtraCalldepth int `json:"extra_calldepth"`
}

var vLog *VilLog

/**
 * 日志 控制结构体
 */
type VilLog struct{
	logConf *VilLogConf
	log *logging.Logger
	logformat logging.Formatter
}

/**
 * 初始化
 */
func init(){
	//日志模块默认配置
	conf := &VilLogConf{
		ConfFilePath:"",
		Path:"./log",
		Level:"DEBUG",
		FormatMd: 2,
		FileBackEnd:true,
		StderrBackEnd:false,
		NetBackEnd:false,
		ModuleName:"vil",
		ExtraCalldepth:2,
	}
	vLog = NewVilLog(conf)
}

/**
 * 获取配置 实例
 */
func NewLogConfig(fileName string) (*VilLogConf, error){
	var logconf = new(VilLogConf)

	if err := LoadJsonData(fileName, logconf); err != nil{
		return nil, err;
	}
	return logconf, nil
}

/**
 * 获取 日志对象
 */
func NewVilLog(logconf *VilLogConf) (*VilLog){
	var villog = &VilLog{}
	if logconf == nil{
		villog.logConf  = &VilLogConf{
			ConfFilePath:"",
			Path:"./log",
			Level:"INFO",
			FormatMd: 2,
			FileBackEnd:true,
			StderrBackEnd:false,
			NetBackEnd:false,
			ModuleName:"vil",
			ExtraCalldepth:2,
		}
	}else{
		villog.logConf = logconf;
	}
	villog.logformat = logging.MustStringFormatter(LogFormat[villog.logConf.FormatMd])
	// 获取 go-logging 实例
	villog.log = logging.MustGetLogger(villog.logConf.ModuleName)
	villog.AddLogBackend()
	return villog
}

/**
 * 添加日志输出的终端
 */
func (self VilLog)AddLogBackend()(err error){
	var backend logging.LeveledBackend
	if self.logConf.FileBackEnd {
		//判断是否存在该文件夹
		err := os.MkdirAll(self.logConf.Path, 0777)
		if err != nil {
			fmt.Println("mkdir error:", err)
			panic(0)
		}
		// 打开一个文件
		file, err := os.OpenFile(path.Join(self.logConf.Path, self.logConf.ModuleName + "_info.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil{
			return err
		}

		backend = self.getLogBackend(file, self.logformat, LogLevelMap[self.logConf.Level]);
	}else {
		backend = self.getLogBackend(os.Stderr, self.logformat, LogLevelMap[self.logConf.Level]);
	}
	self.log.ExtraCalldepth = self.logConf.ExtraCalldepth
	self.log.SetBackend(backend)
	return nil
}

/**
 * 获取终端
 */
func (self VilLog)getLogBackend(out io.Writer, format logging.Formatter, level int)( logging.LeveledBackend){
	backend := logging.NewLogBackend(out, "", 1)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.Level(level), "")
	return backendLeveled
}

func (self VilLog)LogINFO(infmt string,args ... interface{}){
	self.log.Infof(infmt, args ...)
}

func (self VilLog)LogERRO(infmt string,args ... interface{}){
	self.log.Errorf(infmt, args ...)
}

func (self VilLog)LogDEBU(infmt string,args ... interface{}){
	self.log.Debugf(infmt, args ...)
}

func (self VilLog)LogWARN(infmt string,args ... interface{}){
	self.log.Warningf(infmt, args ...)
}

func LogI(infmt string,args ... interface{}){
	vLog.LogINFO(infmt, args ... )
}

func LogE(infmt string,args ... interface{}){
	vLog.LogERRO(infmt, args ... )
}

func LogD(infmt string,args ... interface{}){
	vLog.LogDEBU(infmt, args ... )
}

func LogW(infmt string,args ... interface{}){
	vLog.LogWARN(infmt, args ... )
}
