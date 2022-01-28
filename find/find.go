package find

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

//定义 json 格式的结构体
type Know struct {
	Data []map[string]interface{} `json:"data"`
}

type LogStruct struct {
	Knowledge Know `json:"knowledge"`
}

func Read(production string) []byte {
	fileName := "knowledge/" + production
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("read file failed", err)
	}
	return file
}

func Logcheck(production, match_rate, pass_num string) {
	//解析 knowledge 知识库文件里的字段信息
	file := Read(production)
	var know LogStruct
	err := json.Unmarshal(file, &know)
	if err != nil {
		fmt.Println(err)
	}
	data := know.Knowledge.Data
	for _, v := range data {
		//定义存储 问题-答案 知识库
		var Questings []string //问题
		var Answers []string   //答案
		var Mark []string      //知识库唯一编号
		//模块名称 和 日志路径 先放在数组，这样的话 答案跟问题 就差了 2 个 index(问题 index-2 = 答案 index)
		projectModeName := v["metric"].(map[string]interface{})["name"].(string)
		logdir := v["metric"].(map[string]interface{})["logdir"] //需要到 shell 里解析成字符串
		//将 knowledge json 文件里的带$环境变量路径进行解析
		command := "echo " + logdir.(string)
		logdirShell := exec.Command("bash", "-c", command)
		logdirReall, _ := logdirShell.CombinedOutput()
		logdirFinal := strings.Replace(string(logdirReall), "\n", "", -1) //这里有个换行符需要替换，否则下面 os.open 一直报错

		//调用 scp 函数，给远程日志拉取到本地
		DiskCheck() //拉取前检查磁盘，大于 80% 停止程序
		LogName := ScpLog(production, projectModeName, logdirFinal)

		Questings = append(Questings, projectModeName, logdirFinal)
		values := v["values"]
		list := values.([]interface{})
		for _, v := range list {
			value := v.([]interface{})
			//通过空接口断言取值
			for k, v := range value {
				if k == 0 {
					Mark = append(Mark, v.(string)) //空接口断言取值
				}
				if k == 1 {
					Questings = append(Questings, v.(string))
				} else if k == 2 {
					Answers = append(Answers, v.(string))
				}
			}
		}
		//fmt.Printf("%s", Questings)
		MatchRate, _ := strconv.ParseFloat(match_rate, 32)
		PassNum, _ := strconv.ParseInt(pass_num, 10, 8)
		LogMatching(Questings, Answers, Mark, LogName, MatchRate, PassNum)
	}
}

func LogMatching(Questing, Answers, Mark []string, LogName []string, MatchRate float64, PassNum int64) {
	//打开 日志文件 地址
	for _, v := range LogName {
		file, err := os.Open("log/" + v)
		if err != nil {
			fmt.Println("ERROR EXIT:", err)
		}
		//输出日志文件路径信息
		//	fmt.Printf("\n日志【%s】与知识库匹配结果如下:\n", v)
		Color("GREEN", "\n日志【")
		Color("GREEN", v)
		Color("GREEN", "】与知识库匹配结果如下:\n")
		Color("YELLOW", "------------------------------------------------------------------------------------------------------------------------\n")

		defer file.Close()
		reader := bufio.NewReader(file)

		//定义 匹配到的 日志列表
		var matches []string
		LineNum := 0  //定义日志文件的 行号
		MatchNum := 0 //定义初次匹配到的个数
		//循环读取文件的一行，进行 知识库 匹配
		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			LineNum++
			//fmt.Println("line_startnum:", LineNum, "日志内容:", string(line))
			log := []string{string(line)}
			var checkOut []string
			//对读取到的这行数据，进行 知识库 循环匹配
			for i := 2; i < len(Questing); i++ {
				var weight string
				//fmt.Println("知识库行号:", i-1)
				if Questing[i] == "ERROR" {
					matchOrNo := strings.Contains(log[0], "ERROR")
					if matchOrNo == true {
						checkOut = log
						weight = "0" //降低权重
					}
				} else if Questing[i] == "Exception" {
					matchOrNo := strings.Contains(log[0], "Exception")
					if matchOrNo == true {
						checkOut = log
						weight = "0" //降低权重
					}
				} else {
					checkOut = fuzzy.Find(Questing[i], log)
					weight = "1" //加权处理
				}
				//日志与知识库进行匹配
				if checkOut != nil {
					MatchNum += 1 //说明匹配到了
					//匹配到 的日志 进行 去重
					count := 0
					//加权参数加入
					checkOut = append(checkOut, weight)
					//fmt.Println("i come here", "line num:", checkOut[1])
					for i := 0; i < len(matches); i = i + 2 {
						if checkOut[0] != matches[i] {
							//fmt.Println("\n", "日志行号:", LineNum, "匹配库里的行号:", i)
							//fmt.Println("checkOut0:", checkOut[0], "\t", "match[i]:", matches[i])
							percent := Screen(checkOut[0], matches[i], PassNum)
							if percent < MatchRate {
								count = count + 2 //对每一条匹配到的日志进行数组循环比较，不在的话，计数器+1，如全部不在，才算是一条新的报错信息，最后入库输出。
								//fmt.Println("count:", count)
							} else {
								//fmt.Println("jbbbbbb", "check[1]:", checkOut[1])
								if checkOut[1] == "1" {
									if matches[i+1] == "0" {
										count = count + 2
										//fmt.Println("权重取代")
									}
								}
							}
						}
					}
					if count == len(matches) {
						matches = append(matches, checkOut...)
						//fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
						fmt.Printf("\n【第 %d 处报错】: 错误代码[%s]\n", (count+2)/2, Mark[i-2])
						fmt.Printf("【日志信息】:[第 %d 行]==>【%s】\n", LineNum, checkOut[0])
						fmt.Printf("【知识库解决方案】: 【%s】\n\n", Answers[i-2])
					}

				}
			}
		}
		//fmt.Println("【与知识库匹配到日志个数为:", MatchNum, "】")
		Color("YELLOW", "【 与知识库匹配到日志个数为: ")
		Color("YELLOW", strconv.Itoa(MatchNum))
		Color("YELLOW", " 】\n")
		//	fmt.Println("【经过相似度去重后匹个数为:", len(matches)/2, "】")
		Color("YELLOW", "【 经过相似度去重后匹个数为: ")
		Color("YELLOW", strconv.Itoa(len(matches)/2))
		Color("YELLOW", " 】\n")
	}
}
