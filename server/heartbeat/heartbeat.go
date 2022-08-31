package heartbeat

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/net/context"
	"kp-runner/log"
	gonet "net"
	"strings"
	"time"
)

var (
	heartbeat = new(HeartBeat)
	LocalIp   = ""
)

func CheckHeartBeat(ctx context.Context) *HeartBeat {
	heartbeat.name = GetHostName()
	heartbeat.cpu = GetCpuUsed()
	heartbeat.mem = GetMemUsed()
	heartbeat.cpuLoad = GetCPULoad()
	heartbeat.network = GetNetwork("")
	heartbeat.disk = GetDiskUsed("")
	return heartbeat
}

type HeartBeat struct {
	ip      string
	name    string
	cpu     float64
	cpuLoad *load.AvgStat
	mem     float64
	network []uint64
	disk    float64
}

// CPU信息

func GetCpuUsed() float64 {
	percent, _ := cpu.Percent(time.Second, false) // false表示CPU总使用率，true为单核
	return percent[0]
}

// 负载信息

func GetCPULoad() (info *load.AvgStat) {
	info, _ = load.Avg()
	return
}

// 内存信息

func GetMemUsed() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return memInfo.UsedPercent
}

// 主机信息

func GetHostName() string {
	hostInfo, _ := host.Info()
	return hostInfo.Hostname
}

// 磁盘信息

func GetDiskUsed(path string) float64 {
	parts, _ := disk.Partitions(true)
	for _, part := range parts {
		partInfo, _ := disk.Usage(part.Mountpoint)
		if partInfo.Path == path {
			return partInfo.UsedPercent
		}
	}
	return 0
}

// 网络信息

func GetNetwork(networkName string) []uint64 {
	var network []uint64
	netIOs, _ := net.IOCounters(true)
	for _, netIO := range netIOs {
		if netIO.Name == networkName {
			network = append(network, netIO.BytesSent)
			network = append(network, netIO.BytesRecv)
			return network
		}
	}
	return network
}

func InitLocalIp() {

	conn, err := gonet.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Logger.Error("udp服务：", err)
		return
	}
	localAddr := conn.LocalAddr().(*gonet.UDPAddr)
	LocalIp = strings.Split(localAddr.String(), ":")[0]
	log.Logger.Info("本机ip：", LocalIp)
}
