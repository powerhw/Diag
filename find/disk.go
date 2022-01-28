package find

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DiskCheck() {
	command := "df . -h | awk {'print $5'} | grep -v Use | sed 's/%//g'"
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("disck check faild:\n%s\n", string(out))
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	Diskuse := strings.Replace(string(out), "\n", "", -1)
	DiskUse, _ := strconv.Atoi(Diskuse)

	if DiskUse > 80 {
		fmt.Printf("危险: 当前磁盘使用率 %d%%,请换个挂载点执行本程序！", DiskUse)
		os.Exit(1)
	}
}
