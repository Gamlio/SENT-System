//go:build windows

package collector

import (
	"context"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// Collect implements a lightweight ETW-like polling fallback on Windows.
// It gathers recent process creations, established outbound network flows, and simulated registry events.
func (s *ETWSensor) Collect() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var procEvents []ETWProcessEvent
	var netEvents []ETWNetworkEvent
	var regEvents []ETWRegistryEvent
	var fileEvents []ETWFileIOEvent

	// 1. Filter: Process Creation Events
	if s.Filter.EnableProcess {
		procs, err := GetProcessesCached()
		if err == nil {
			now := time.Now().Unix()
			for _, p := range procs {
				select {
				case <-ctx.Done():
					break
				default:
				}

				// Try to get create time; skip errors silently
				if ct, err := p.CreateTime(); err == nil {
					// CreateTime returns milliseconds since epoch
					createSec := int64(ct / 1000)
					// consider "recent" processes within last 60 seconds as process-start events
					if now-createSec <= 60 {
						name, _ := p.Name()
						exe, _ := p.Exe()
						cmd, _ := p.Cmdline()
						procEvents = append(procEvents, ETWProcessEvent{
							PID:        p.Pid,
							Name:       name,
							ExePath:    exe,
							Cmdline:    cmd,
							CreateTime: createSec,
						})
					}
				}
			}
		} else {
			log.Printf("[ETW] Lỗi lấy tiến trình: %v", err)
		}
	}

	// 2. Filter: Network Connection Events
	if s.Filter.EnableNetwork {
		if conns, err := net.Connections("inet"); err == nil {
			for _, c := range conns {
				// Only consider established outbound with remote address
				if c.Status == "ESTABLISHED" && c.Raddr.IP != "" {
					// Map to fields; PID may be 0 if not available
					netEvents = append(netEvents, ETWNetworkEvent{
						PID:        c.Pid,
						Name:       "", // name optional; could map via process table if needed
						RemoteIP:   c.Raddr.IP,
						RemotePort: uint32(c.Raddr.Port),
						Proto:      networkProtoString(c.Type),
						Timestamp:  time.Now().Unix(),
					})
				}
			}
		}
	}

	// 3. Filter: Registry Modification Events
	if s.Filter.EnableRegistry {
		// Note: Polling Registry changes via gopsutil is not supported natively.
		// In a real kernel-driver scenario, we would stream ETW registry events here.
		// For now, we prepare the structure so that the backend can ingest it when the driver is ready.
		// regEvents = append(regEvents, ETWRegistryEvent{...})
	}

	// 4. Filter: File IO Events
	if s.Filter.EnableFileIO {
		fileEvents = collectFileIOEvents(ctx)
	}

	payload := map[string]interface{}{
		"type":      "etw_polling",
		"processes": procEvents,
		"network":   netEvents,
		"registry":  regEvents,
		"file_io":   fileEvents,
	}

	return payload, nil
}

func networkProtoString(proto uint32) string {
	switch proto {
	case 1:
		return "tcp"
	case 2:
		return "udp"
	case 3:
		return "icmp"
	default:
		return "unknown"
	}
}

func collectFileIOEvents(ctx context.Context) []ETWFileIOEvent {
	// Placeholder for future native File IO / ETW file operation capture.
	// Windows file IO tracking typically requires kernel-mode ETW providers or driver support.
	select {
	case <-ctx.Done():
		return []ETWFileIOEvent{}
	default:
		return []ETWFileIOEvent{}
	}
}
