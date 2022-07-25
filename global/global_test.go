package global

import (
	"fmt"
	"kp-controller/config"
	"testing"
)

func TestGetRunnerServers(t *testing.T) {
	servers := GetRunnerServers(config.Config["server.runner.name"].(string))
	fmt.Println("servers:", servers)
}
