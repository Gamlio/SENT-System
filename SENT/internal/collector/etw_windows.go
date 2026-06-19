package collector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
)

var (
	etwSession    *etw.RealTimeSession
	etwEvents     []interface{}
	etwMutex      sync.Mutex
	etwOnce       sync.Once
	isRealtimeETW bool
)

const MaxEventBuffer = 10000

const (
	KernelProcessProvider  = "{22fb2349-c442-4f0d-a174-88f33e5078c3}"
	KernelNetworkProvider  = "{7dd42a49-5329-4832-8dfd-43d979153a88}"
	KernelRegistryProvider = "{ae53734e-7b3b-4560-9a3d-424a1801804f}"
	KernelFileIOProvider   = "{ed54d105-882b-4e68-b715-bf6724632050}"
)

func startRealtimeETW(filter EventFilter) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ETW] Panic khi khởi tạo Session: %v. Đảm bảo chạy với quyền Admin/SYSTEM.", r)
			isRealtimeETW = false
		}
	}()

	etwSession = etw.NewRealTimeSession("SENT_EDR_Realtime_Trace")

	if filter.EnableProcess {
		_ = etwSession.EnableProvider(etw.Provider{GUID: KernelProcessProvider})
	}
	if filter.EnableNetwork {
		_ = etwSession.EnableProvider(etw.Provider{GUID: KernelNetworkProvider})
	}
	if filter.EnableRegistry {
		_ = etwSession.EnableProvider(etw.Provider{GUID: KernelRegistryProvider})
	}
	if filter.EnableFileIO {
		_ = etwSession.EnableProvider(etw.Provider{GUID: KernelFileIOProvider})
	}

	consumer := etw.NewRealTimeConsumer(context.Background())
	consumer.FromSessions(etwSession)

	go func() {
		for e := range consumer.Events {
			processSingleEvent(e)
		}
	}()

	isRealtimeETW = true
	log.Println("[ETW] Chế độ giám sát Real-time Kernel Event Tracing đã sẵn sàng.")
}

func processSingleEvent(e *etw.Event) {
	etwMutex.Lock()
	if len(etwEvents) >= MaxEventBuffer {
		etwMutex.Unlock()
		return
	}
	etwMutex.Unlock()

	providerID := "{" + strings.ToLower(strings.Trim(e.System.Provider.Guid, "{}")) + "}"
	var eventToAppend interface{}
	eventTime := time.Now().Unix()

	switch providerID {
	case KernelProcessProvider:
		if e.System.EventID == 1 {
			evt := ETWProcessEvent{
				PID:        int32(e.System.Execution.ProcessID),
				Name:       "ProcessStart",
				CreateTime: eventTime,
			}
			if v, ok := e.EventData["ImageName"]; ok {
				evt.ExePath = fmt.Sprintf("%v", v)
			}
			if v, ok := e.EventData["CommandLine"]; ok {
				evt.Cmdline = fmt.Sprintf("%v", v)
			}
			eventToAppend = evt
		}

	case KernelNetworkProvider:
		if e.System.EventID == 10 {
			evt := ETWNetworkEvent{
				PID:       int32(e.System.Execution.ProcessID),
				Timestamp: eventTime,
				Proto:     "tcp",
			}
			if v, ok := e.EventData["daddr"]; ok {
				evt.RemoteIP = fmt.Sprintf("%v", v)
			}
			if v, ok := e.EventData["dport"]; ok {
				fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &evt.RemotePort)
			}
			eventToAppend = evt
		}

	case KernelRegistryProvider:
		if e.System.EventID == 1 || e.System.EventID == 5 {
			action := "modify"
			if e.System.EventID == 1 {
				action = "create"
			}
			evt := ETWRegistryEvent{
				PID:       int32(e.System.Execution.ProcessID),
				Name:      fmt.Sprintf("Registry-ID-%d", e.System.EventID),
				Action:    action,
				Timestamp: eventTime,
			}
			if v, ok := e.EventData["KeyName"]; ok {
				evt.KeyPath = fmt.Sprintf("%v", v)
			}
			eventToAppend = evt
		}

	case KernelFileIOProvider:
		if e.System.EventID == 64 || e.System.EventID == 68 {
			op := "read/open"
			if e.System.EventID == 68 {
				op = "write"
			}
			evt := ETWFileIOEvent{
				PID:       int32(e.System.Execution.ProcessID),
				Name:      fmt.Sprintf("FileIO-ID-%d", e.System.EventID),
				Operation: op,
				Timestamp: eventTime,
			}
			if v, ok := e.EventData["FileName"]; ok {
				evt.FilePath = fmt.Sprintf("%v", v)
			}
			eventToAppend = evt
		}
	}

	if eventToAppend != nil {
		etwMutex.Lock()
		if len(etwEvents) < MaxEventBuffer {
			etwEvents = append(etwEvents, eventToAppend)
		}
		etwMutex.Unlock()
	}
}

func (s *ETWSensor) Collect() (interface{}, error) {
	etwOnce.Do(func() {
		startRealtimeETW(s.Filter)
	})

	if !isRealtimeETW {
		return nil, fmt.Errorf("ETW Real-time chưa được kích hoạt (yêu cầu quyền Admin)")
	}

	etwMutex.Lock()
	collected := etwEvents
	etwEvents = make([]interface{}, 0)
	etwMutex.Unlock()

	var procEvents []ETWProcessEvent
	var netEvents []ETWNetworkEvent
	var regEvents []ETWRegistryEvent
	var fileEvents []ETWFileIOEvent

	for _, ev := range collected {
		switch v := ev.(type) {
		case ETWProcessEvent:
			procEvents = append(procEvents, v)
		case ETWNetworkEvent:
			netEvents = append(netEvents, v)
		case ETWRegistryEvent:
			regEvents = append(regEvents, v)
		case ETWFileIOEvent:
			fileEvents = append(fileEvents, v)
		}
	}

	payload := map[string]interface{}{
		"type":      "etw_realtime",
		"processes": procEvents,
		"network":   netEvents,
		"registry":  regEvents,
		"file_io":   fileEvents,
	}

	return payload, nil
}
