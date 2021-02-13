package tests

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/remoteit/systemkit-processes/contracts"
	procMon "github.com/remoteit/systemkit-processes/monitor"
)

func TestAbs(t *testing.T) {
	wg := sync.WaitGroup{}

	udpPortG := 0

	processTag := "aaaa"

	monitor := procMon.New()
	monitor.SpawnWithTag(contracts.ProcessTemplate{
		Executable: "/usr/bin/connectd",
		Args: []string{
			"-s", "-mfg", "33280", "-ptf", "256", "-p", "bmljb2xhZUByZW1vdGUuaXQ=", "EA5AE177DCAB4A7329A853B84DF7689D6B8602EE", "80:00:00:00:01:0A:01:BA", "T30002", "2", "1.1.1.1", "0.0.0.0", "35", "0", "0",
		},
		StdoutReader: func(params interface{}, outputData []byte) {
			rawDataLine := string(outputData)

			udpPortStr := "bound to UDP port"
			if strings.Contains(rawDataLine, udpPortStr) {

				rawDataLine = rawDataLine[strings.Index(rawDataLine, udpPortStr)+len(udpPortStr):]
				rawDataLine = strings.TrimSpace(rawDataLine)

				udpPort, err := strconv.Atoi(rawDataLine)
				if err == nil {
					udpPortG = udpPort
				}
			}
		},
		OnStopped: func(params interface{}) {
			fmt.Println("STOPPED !!!")
		},
	}, processTag)

	wg.Add(1)
	go func() {
		time.Sleep(5 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("SendKillPacket !!!")
	SendKillPacket(udpPortG)

	wg.Add(1)
	go func() {
		time.Sleep(5 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("StopWithTimeout !!!")
	monitor.StopWithTimeout(processTag, 10, 500*time.Millisecond)

	wg.Add(1)
	go func() {
		time.Sleep(5 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	fmt.Println("DONE !!!")
}

func SendKillPacket(udpPort int) {
	if udpPort < 1 {
		return
	}

	connectdKillMessage := []byte{
		0x00, 0x00, // spi
		0x00, 0x00, // spi
		0x00, 0x00, // salt
		0x00, 0x00, // salt
		0x00, 0x40, // shutdown
		0x00, 0x00, // source
		0x00, 0x00, // data terminator
	}

	conn, err := net.Dial("udp", fmt.Sprintf(`127.0.0.1:%d`, udpPort))
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write(connectdKillMessage)
}
