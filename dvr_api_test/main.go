package main

const (
	DEVICE_SVR_ENDPOINT string = "192.168.1.77:9047"
)

func main() {
	go func() {
		err := SimulateDevices()
	}()
}
