package heartbeat

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"testing"
)

func TestGetCpuInfo(t *testing.T) {
	//ctx := context.Background()
	//heartbeat = CheckHeartBeat(ctx)
	//by, _ := json.Marshal(heartbeat)
	//fmt.Println(string(by))
}

//
//func TestGetHostInfo(t *testing.T) {
//	GetHostInfo()
//}
//
//func TestGetMemInfo(t *testing.T) {
//	GetMemInfo()
//}
//
func TestGetDiskInfo(t *testing.T) {
	a, _ := disk.IOCounters()
	for k, _ := range a {
		b, _ := disk.Usage(k)
		c, _ := json.Marshal(b)
		fmt.Println(":    ", string(c))
	}
}

//
func TestGetNetInfo(t *testing.T) {
	infos, _ := net.IOCounters(true)
	fmt.Println(infos)
}
