package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

//定义采集 prometheus 指标 以及 阈值
type Metrics struct {
	Name, Url, Unit, StarTime, EndTime string
	Limit                              float64
}

//根据命令行输入的参数生产 prometheus api url
func Promql(start_time, end_time, s_time, e_time, step string) {
	//Baseurl := "http://10.10.25.8:9090/api/v1/query_range?"
	Baseurl := "http://localhost:8310/api/v1/query_range?"

	//cpu 使用率,单位 % (已经乘了100)
	CpuUrl := Baseurl + "query=(1-avg(rate(node_cpu_seconds_total{mode='idle'}[5m]))by(instance))*100&start=" + start_time + "&end=" + end_time + "&step=" + step
	CpuMetric := Metrics{Name: "Cpu 使用率", Unit: "%", Limit: 90, Url: CpuUrl, StarTime: s_time, EndTime: e_time}
	//内存使用占比，单位 % (已经乘了100)
	MemUse := Baseurl + "query=((node_memory_MemTotal_bytes)-(node_memory_MemFree_bytes)-(node_memory_Cached_bytes)-(node_memory_Buffers_bytes))/(node_memory_MemTotal_bytes)*100&start=" + start_time + "&end=" + end_time + "&step=" + step
	MemMetric := Metrics{Name: "Mem 使用率", Unit: "%", Limit: 85, Url: MemUse, StarTime: s_time, EndTime: e_time}
	//磁盘 iops
	Iops := Baseurl + "query=rate(node_disk_reads_completed_total{}[5m])&start=" + start_time + "&end=" + end_time + "&step=" + step
	IopsMetric := Metrics{Name: "磁盘 Iops", Unit: " ops/sec", Limit: 10000, Url: Iops, StarTime: s_time, EndTime: e_time}
	//每一秒内 I/O 操作占时
	IopsSec := Baseurl + "query=rate(node_disk_io_time_seconds_total{}[5m])&start=" + start_time + "&end=" + end_time + "&step=" + step
	IopsSecMetric := Metrics{Name: "磁盘 I/O 操作占时", Unit: "%", Limit: 85, Url: IopsSec, StarTime: s_time, EndTime: e_time}
	//网卡流量,单位 mb
	NetDown := Baseurl + "query=rate(node_network_receive_bytes_total{}[5m])*8/1000/1000&start=" + start_time + "&end=" + end_time + "&step=" + step
	NetDownMetric := Metrics{Name: "网卡下载流量", Unit: "mb", Limit: 10000, Url: NetDown, StarTime: s_time, EndTime: e_time}
	NetPut := Baseurl + "query=rate(node_network_transmit_bytes_total{}[5m])*8/1000/1000&start=" + start_time + "&end=" + end_time + "&step=" + step
	NetPutMetric := Metrics{Name: "网卡上传流量", Unit: "mb", Limit: 10000, Url: NetPut, StarTime: s_time, EndTime: e_time}

	Parsing(CpuMetric)
	Parsing(MemMetric)
	Parsing(IopsMetric)
	Parsing(IopsSecMetric)
	Parsing(NetDownMetric)
	available := Parsing(NetPutMetric)

	//如果指标值为空，返回提示
	if available == 0 {
		fmt.Println("no available data,please submit a query condition again!")
	}
}

//通过 api url 获取 prometheus 数据
func Parsing(Query Metrics) (ava int) {

	type Data struct {
		Values []map[string]interface{} `json:"result"`
	}

	type Person struct {
		Data Data `json:"data"`
	}

	resp, err := http.Get(Query.Url)
	if err != nil {
		fmt.Println("get failed, err:", err)
		return
	}
	defer resp.Body.Close()
	b, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		fmt.Println("read from resp.Body failed,err:", err)
		return
	}

	var p Person
	err2 := json.Unmarshal(b, &p)
	if err2 != nil {
		fmt.Println(err2)
	}
	//定义个指标，没数据的时候提示 no available data
	available := 0

	//获取 instance，values 数据
	data := p.Data.Values
	for _, v := range data {
		instance := v["metric"].(map[string]interface{})["instance"]
		device := v["metric"].(map[string]interface{})["device"]
		values := v["values"]
		available = available + 1 //计算下，与初始值 0 的区别
		var PromeTime []float64
		var PromeData []float64
		list := values.([]interface{})
		for _, v := range list {
			value := v.([]interface{})
			for _, v := range value {
				switch data := v.(type) {
				case float64:
					//debug
					PromeTime = append(PromeTime, data)
				case string:
					//debug
					//fmt.Println("v:", data)
					new1, _ := strconv.ParseFloat(data, 64)
					PromeData = append(PromeData, float64(new1))
				}
			}
		}
		//输出指标
		if device == nil {
			fmt.Printf("\n\n%c[0;30;32m【主机 %s %s】%c[0m\n", 0x1B, instance, Query.Name, 0x1B)
			//fmt.Println("\n【 主机", instance, "在", Query.StarTime, "->", Query.EndTime, "这段时间", Query.Name, "】")
		} else {
			//fmt.Println("\n【 主机", instance, "在", Query.StarTime, "->", Query.EndTime, "这段时间", device, Query.Name, "】")
			fmt.Printf("\n\n%c[0;30;32m【主机 %s [%s] %s】%c[0m\n", 0x1B, instance, device, Query.Name, 0x1B)
		}
		for i := 0; i < len(PromeData); i++ {
			//格式化时间输出 float64->int64->string
			TimeOut := time.Unix(int64(PromeTime[i]), 0).Format("15:04.05")
			//超过阈值的标记颜色输出
			if PromeData[i] > Query.Limit {
				fmt.Printf("%c[4;30;31m[[ %s => %.1f%s ]]%c[0m\n", 0x1B, TimeOut, PromeData[i], Query.Unit, 0x1B)
			} else {
				fmt.Printf("[[ %s => %.1f%s ]]", TimeOut, PromeData[i], Query.Unit)
			}
		}
	}

	return available
}

//颜色函数
//fmt.Printf("\n\n %c[1;31;43m%s%c[0m", 0x1B, out, 0x1B)
