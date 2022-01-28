package find

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func sTz(a []string) string {
	var out string
	for _, i := range a {
		out = out + " " + i
	}
	return out
}

func stringIorN(a []string, b string, num int64) []string {
	var time int64
	for k := 1; k < len(a); k = k + 1 {
		matchOrNo := strings.Contains(b, a[k])
		if matchOrNo == false {
			time = time + 1
			a = append(a[:k], a[k+1:]...)
			k = k - 1        //数组 a 长度经过上面去除后会变短一个
			if time >= num { //限制去除的个数
				break
			}
		}
	}
	return a
}

func Screen(orgLog, referLog string, pass_num int64) float64 {
	//a := "2021-12-27T10:55:44.406+0800 INFO - Thread-23 [] c.s.p.d.w.k.operator.KafkaConsumerOperator []: Terminator really shutdown now. reason=[reason:[get job status failed],exception_stack=[org.apache.hadoop.fs.FSError: java.io.IOException: No space left on device"
	//b := "2021-12-27T10:55:44.406+0800 INFO - Thread-24 [] c.s.p.d.w.k.operator.KafkaConsumerOperator []: Terminator really shutdown now. reason=[reason:[get job status failed],exception_stack=[org.apache.hadoop.fs.FSError: java.io.IOException: No space left on device"
	var result_a, result_e, result_m []string
	var all float64
	new_w := strings.Fields(orgLog)
	new_a := new_w
	//if len(new_w) > 3 {
	new_a = stringIorN(new_w, referLog, pass_num)
	//}
	//fmt.Println("new_a", new_a)
	//fmt.Println("refer:", referLog)
	half := len(new_a) / 2
	//for k, v := range new_a {
	//	fmt.Println(k, v)
	//}
	for i := 1; i < half+1; i++ {
		//fmt.Println("i: ", i)
		result_a = new_a[i:]
		result_e = new_a[:len(new_a)-i]
		result_m = new_a[i : len(new_a)-i]

		//debug
		//数组变字符串
		result_1 := sTz(result_a)
		result_2 := sTz(result_e)
		result_3 := sTz(result_m)

		//fmt.Println("+1:", result_1)
		//fmt.Println("+2:", result_2)
		//fmt.Println("+3:", result_3)

		match_2 := fuzzy.Match(result_1, referLog)
		if match_2 == true {
			all = all + 1
			//fmt.Println("begin: ", all, result_1)
			continue
		}

		match_3 := fuzzy.Match(result_2, referLog)
		if match_3 == true {
			all = all + 1
			//fmt.Println("end: ", all, result_2)
			continue
		}
		match_1 := fuzzy.Match(result_3, referLog)
		if match_1 == true {
			all = all + 1
			//fmt.Println("media: ", all, result_3)
			continue
		}
	}
	per := (all / float64(half)) * 100
	//fmt.Println("all->zhong", half+1, all, "per:", per)
	return per
}
