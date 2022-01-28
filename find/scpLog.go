package find

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ScpLog(projectName, projecModeName, Logdir string) []string {
	var LogName []string
	switch projectName {
	case "sdf":
		LogName = SdfLog(projectName, projecModeName, Logdir)
	case "sa":
		LogName = SaLog(projectName, projecModeName, Logdir)
	}
	return LogName
}

func SaLog(projectName, projecModeName, Logdir string) []string {
	var HostFile string
	var Logname []string
	command := "spadmin status -m web -p sa 2>&1|egrep -v \"INFO|product\"|sed 's/|//g'|sed 's/ //g' > log/host.txt"
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	//读取文件
	file, err := os.Open("log/host.txt")
	if err != nil {
		fmt.Println("ERROR EXIT:", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	//循环读取文件的一行
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error:read host file error!")
		}
		i := strings.Trim(string(line), "\n")
		HostFile = HostFile + string(i)
	}
	//筛选出各个角色对应的主机
	spit := strings.Split(HostFile, ":")
	for _, v := range spit {
		var logfile string
		if strings.Contains(v, "@") {
			HostWeb := strings.Split(v, "@")[1]
			logfile = projecModeName + ".log_" + HostWeb
			SshCommand := "scp " + HostWeb + ":" + Logdir + " log/" + logfile
			SshCmd := exec.Command("bash", "-c", SshCommand)
			out1, err := SshCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("combined out:\n%s\n", string(out1))
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}

		}
		if logfile != "" {
			Logname = append(Logname, logfile)
		}
	}
	return Logname
}

func SdfLog(projectName, projecModeName, Logdir string) []string {
	var HostMaster, HostKc, HostKtp, HostMain, HostPm string
	var HostFile string
	var SshCommand string
	var LogName []string

	command := "spadmin status -m data_loader -p sdf 2>&1|egrep -v \"INFO|product\"|sed 's/|//g'|sed 's/ //g' > log/host.txt"
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("combined out:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	//读取文件
	file, err := os.Open("log/host.txt")
	if err != nil {
		fmt.Println("ERROR EXIT:", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	//循环读取文件的一行
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error:read host file error!")
		}
		i := strings.Trim(string(line), "\n")
		HostFile = HostFile + string(i)
	}
	//筛选出各个角色对应的主机
	var Logname string
	spit := strings.Split(HostFile, ":")
	for _, v := range spit {
		if strings.Contains(v, "!") {
		} else {
			if strings.Contains(v, "master") {
				HostMaster = strings.Split(v, ":")[0]
				HostMaster = strings.Split(HostMaster, "@")[1]
			} else if strings.Contains(v, "worker_kafka_consume") {
				HostKc = strings.Split(v, ":")[0]
				HostKc = strings.Split(HostKc, "@")[1]
			} else if strings.Contains(v, "worker_kudu_to_parquet") {
				HostKtp = strings.Split(v, ":")[0]
				HostKtp = strings.Split(HostKtp, "@")[1]
			} else if strings.Contains(v, "worker_maintenance") {
				HostMain = strings.Split(v, ":")[0]
				HostMain = strings.Split(HostMain, "@")[1]
			} else if strings.Contains(v, "worker_project_manager") {
				HostPm = strings.Split(v, ":")[0]
				HostPm = strings.Split(HostPm, "@")[1]
			}
		}
		if projecModeName == "master" {
			Logname = projecModeName + ".log_" + HostMaster
			SshCommand = "scp " + HostMaster + ":" + Logdir + " log/" + Logname
		} else if projecModeName == "worker_kafka_consumer" {
			Logname = projecModeName + ".log_" + HostKc
			SshCommand = "scp " + HostKc + ":" + Logdir + " log/" + Logname
		} else if projecModeName == "worker_kudu_to_parquet" {
			Logname = projecModeName + ".log_" + HostKtp
			SshCommand = "scp " + HostKtp + ":" + Logdir + " log/" + Logname
		} else if projecModeName == "worker_maintenance" {
			Logname = projecModeName + ".log_" + HostMain
			SshCommand = "scp " + HostMain + ":" + Logdir + " log/" + Logname
		} else if projecModeName == "worker_project_manager" {
			Logname = projecModeName + ".log_" + HostPm
			SshCommand = "scp " + HostPm + ":" + Logdir + " log/" + Logname
		}
	}
	LogName = append(LogName, Logname)
	//fmt.Println("SshCommand:", SshCommand)
	SshCmd := exec.Command("bash", "-c", SshCommand)
	out1, err := SshCmd.CombinedOutput()
	if err != nil {
		//fmt.Printf("combined out:\n%s\n", string(out1))
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		fmt.Println("scp failed", string(out1))
	}
	return LogName
}
