package heartbeat

import (
	"RunnerGo-engine/model"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	gonet "net"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"RunnerGo-engine/config"
	"RunnerGo-engine/log"
	services "RunnerGo-engine/proto"
)

var (
	heartbeat = new(HeartBeat)
	LocalIp   = ""
	Key       = "RunnerMachineList"
)

func CheckHeartBeat() *HeartBeat {
	heartbeat.Name = GetHostName()
	heartbeat.CpuUsage = GetCpuUsed()
	heartbeat.MemInfo = GetMemInfo()
	heartbeat.CpuLoad = GetCPULoad()
	heartbeat.Networks = GetNetwork()
	heartbeat.MaxGoroutines = config.Conf.Machine.MaxGoroutines
	heartbeat.DiskInfos = GetDiskInfo()
	heartbeat.CreateTime = time.Now().Unix()
	heartbeat.ServerType = config.Conf.Machine.ServerType
	heartbeat.CurrentGoroutines = runtime.NumGoroutine()
	return heartbeat
}

func SendHeartBeat(host string, duration int64) {

	ctx := context.TODO()

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		panic(errors.Wrap(err, "cannot load root CA certs"))
	}
	creds := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(creds))

	grpcClient := services.NewKpControllerClient(conn)

	req := new(services.RegisterMachineReq)
	req.IP = LocalIp
	req.Port = config.Conf.Heartbeat.Port
	req.Region = config.Conf.Heartbeat.Region
	ticker := time.NewTicker(time.Duration(duration) * time.Second)
	for {
		select {
		case <-ticker.C:
			_, err = grpcClient.RegisterMachine(ctx, req)
			if err != nil {
				log.Logger.Error("grpc服务心跳发送失败", err)
			}
		}

	}

}

type HeartBeat struct {
	Name              string        `json:"name"`
	CpuUsage          float64       `json:"cpu_usage"`
	CpuLoad           *load.AvgStat `json:"cpu_load"`
	MemInfo           []MemInfo     `json:"mem_info"`
	Networks          []Network     `json:"networks"`
	DiskInfos         []DiskInfo    `json:"disk_infos"`
	MaxGoroutines     int           `json:"max_goroutines"`
	CurrentGoroutines int           `json:"current_goroutines"`
	ServerType        int           `json:"server_type"`
	CreateTime        int64         `json:"create_time"`
}

type MemInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Network struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
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

func GetMemInfo() (memInfoList []MemInfo) {
	memVir := MemInfo{}
	memInfoVir, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoVir.Total
	memVir.Free = memInfoVir.Free
	memVir.Used = memInfoVir.Used
	memVir.UsedPercent = memInfoVir.UsedPercent
	memInfoList = append(memInfoList, memVir)
	memInfoSwap, err := mem.SwapMemory()
	if err != nil {
		return
	}
	memVir.Total = memInfoSwap.Total
	memVir.Free = memInfoSwap.Free
	memVir.Used = memInfoSwap.Used
	memVir.UsedPercent = memInfoSwap.UsedPercent
	memInfoList = append(memInfoList, memVir)
	return memInfoList
}

// 主机信息

func GetHostName() string {
	hostInfo, _ := host.Info()
	return hostInfo.Hostname
}

// 磁盘信息

func GetDiskInfo() (diskInfoList []DiskInfo) {
	disks, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for _, v := range disks {
		diskInfo := DiskInfo{}
		info, err := disk.Usage(v.Device)
		if err != nil {
			continue
		}
		diskInfo.Total = info.Total
		diskInfo.Free = info.Free
		diskInfo.Used = info.Used
		diskInfo.UsedPercent = info.UsedPercent
		diskInfoList = append(diskInfoList, diskInfo)
	}
	return
}

// 网络信息

func GetNetwork() (networkList []Network) {
	netIOs, _ := net.IOCounters(true)
	if netIOs == nil {
		return
	}
	for _, netIO := range netIOs {
		network := Network{}
		network.Name = netIO.Name
		network.BytesSent = netIO.BytesSent
		network.BytesRecv = netIO.BytesRecv
		network.PacketsSent = netIO.PacketsSent
		network.PacketsRecv = netIO.PacketsRecv
		networkList = append(networkList, network)
	}
	return
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

func SendHeartBeatRedis(field string, duration int64) {
	ticker := time.NewTicker(time.Duration(duration) * time.Second)

	for {
		select {
		case <-ticker.C:
			CheckHeartBeat()
			hb, _ := json.Marshal(heartbeat)
			err := model.InsertHeartbeat(Key, field, string(hb))
			if err != nil {
				log.Logger.Error("心跳发送失败, 写入redis失败:   ", err)
			}
		}
	}
}

func SendMachineResources(duration int64) {
	ticker := time.NewTicker(time.Duration(duration) * time.Second)
	key := fmt.Sprintf("MachineMonitor:%s", LocalIp)
	for {
		select {
		case <-ticker.C:
			CheckHeartBeat()
			hb, _ := json.Marshal(heartbeat)
			err := model.InsertMachineResources(key, string(hb))
			if err != nil {
				log.Logger.Error("资源写入失败, 写入redis失败:   ", err)
			}
		}
	}
}
