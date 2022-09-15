package heartbeat

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/proto/app/services"
	gonet "net"
	"strings"
	"time"
)

var (
	heartbeat = new(HeartBeat)
	LocalIp   = ""
	LocalHost = ""
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

func SendHeartBeat(host string, duration int64) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		panic(errors.Wrap(err, "cannot load root CA certs"))
	}
	creds := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, host, grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Logger.Error(fmt.Sprintf("服务注册失败： %s", err))
	}
	defer conn.Close()
	// 初始化grpc客户端

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
