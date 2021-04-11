package cpu_percent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var cpuNum int

func init()  {

	quotafile, err := os.Open("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if err != nil {
		fmt.Println(err)
	}
	periodfile,err := os.Open("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err != nil {
		fmt.Println(err)
	}
	defer quotafile.Close()
	defer periodfile.Close()

	rd := bufio.NewReader(quotafile)
	qutostr,_ ,err  := rd.ReadLine()

	if err != nil {
		fmt.Println(err)
	}
	cpu_quota ,err := strconv.Atoi(string(qutostr))

	period := bufio.NewReader(periodfile)
	periodstr, _, err := period.ReadLine()

	if err != nil {
		fmt.Println(err)
	}
	cpu_period ,err := strconv.Atoi(string(periodstr))

	if err != nil {
		fmt.Println(err)
	}

	if cpu_quota != -1 {

		cpuNum = cpu_quota / cpu_period
	}else {
		cpuNum = runtime.NumCPU()
	}
}

func Percent(interval time.Duration) ([]float64, error) {
	return PercentWithContext(context.Background(), interval)
}

func PercentWithContext(ctx context.Context, interval time.Duration) ([]float64, error) {

	cpuTimes := make([]float64, 0)
	frist := queCpuacct()

	if interval <= 0 {
		interval = time.Second
	}
	time.Sleep(interval)
	last := queCpuacct()
	//防止溢出
	//if last < frist {
	//	frist = queCpuacct()
	//	time.Sleep(interval)
	//	last = queCpuacct()
	//}
	//容器的CPU使用率用总使用率需要除以容器核心数 * 100%，才能和物理机取到的CPU使用率的计算方式一致
	//容器需要考虑超卖的情况
	cpuTimes = append(cpuTimes, (last-frist)/float64(cpuNum))

	return cpuTimes, nil
}

func splitCpuacct(s string) float64 {
	ss := strings.Trim(strings.Split(s, " ")[1], "\n")
	cpuTime, _ := strconv.ParseFloat(ss, 64)
	return cpuTime
}

func queCpuacct() float64 {

	//容器的CPU使用率用总使用率取值文件
	file, err := os.Open("/sys/fs/cgroup/cpu/cpuacct.stat")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	s := make([]string, 0)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		s = append(s, line)
		if err != nil || io.EOF == err {
			break
		}
	}
	userCpuTime := splitCpuacct(s[0])
	sysCpuTime := splitCpuacct(s[1])

	return userCpuTime + sysCpuTime
}

func CPUNum() int  {
	return cpuNum
}



