package main

import (
	"Diag/api"
	"Diag/find"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var start_time, end_time, Step string
var production, match_rate, pass_num string

func main() {
	var cmdPrometheus = &cobra.Command{
		Use:   "prometheus -b '2021-12-05 08:00:00' -e '2021-12-05 11:11:00' -s 5m",
		Short: "Get Prometheus Local Data",
		Run: func(cmd *cobra.Command, args []string) {
			//格式判断，以及默认值设定，不符合格式的程序退出
			StartTime, err := time.ParseInLocation("2006-01-02 15:04:05", start_time, time.Local)
			if err != nil {
				fmt.Println("输入的 start_time 有错误，请按照格式 \"2021-01-01 10:00:00\"")
				os.Exit(1)
			}
			EndTime, err := time.ParseInLocation("2006-01-02 15:04:05", end_time, time.Local)
			if err != nil {
				fmt.Println("输入的 end_time 有错误，请按照格式 \"2021-01-01 10:00:00\"")
				os.Exit(1)
			}
			// int64 -> string
			Stime := strconv.FormatInt(StartTime.Unix(), 10)
			Etime := strconv.FormatInt(EndTime.Unix(), 10)
			api.Promql(Stime, Etime, start_time, end_time, Step)
		},
	}
	var cmdLogcheck = &cobra.Command{
		Use:   "logcheck -p [production]",
		Short: "matching knowledge's error,auto give you answers!",
		Run: func(cmd *cobra.Command, args []string) {
			find.Logcheck(production, match_rate, pass_num)
		},
	}

	var rootCmd = &cobra.Command{
		Use:   "sreadmin",
		Short: "SRE Auxiliary Tool",
	}
	//加载 子命令 prometheus
	rootCmd.AddCommand(cmdPrometheus)
	cmdPrometheus.Flags().StringVarP(&start_time, "begin_time", "b", "2021-12-02 08:00:00", "start_time")
	cmdPrometheus.Flags().StringVarP(&end_time, "end_time", "e", "2021-12-02 10:00:00", "end_time")
	cmdPrometheus.Flags().StringVarP(&Step, "step", "s", "5m", "step")

	//加载 子命令 prometheus
	rootCmd.AddCommand(cmdLogcheck)
	cmdLogcheck.Flags().StringVarP(&production, "production", "p", "sdf", "production name")
	cmdLogcheck.Flags().StringVarP(&match_rate, "match_rate", "m", "80", "matching rate, default 80%")
	cmdLogcheck.Flags().StringVarP(&pass_num, "skip_words", "s", "3", "skip the number of different words")

	rootCmd.Execute()
}
