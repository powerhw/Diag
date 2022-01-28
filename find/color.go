package find

import "fmt"

func Color(color, txt string) {
	switch color {
	case "RED":
		fmt.Printf("%c[1;40;31m%s%c[0m", 0x1B, txt, 0x1B) //红色字体
	case "GREEN":
		fmt.Printf("%c[1;40;32m%s%c[0m", 0x1B, txt, 0x1B) //绿色字体
	case "YELLOW":
		fmt.Printf("%c[1;40;33m%s%c[0m", 0x1B, txt, 0x1B) //黄色字体
	}
}
