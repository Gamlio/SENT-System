package collector

type EventFilter struct {
	EnableProcess  bool
	EnableNetwork  bool
	EnableRegistry bool
	EnableFileIO   bool
}

type ETWSensor struct {
	Filter EventFilter
}

type ETWProcessEvent struct {
	PID        int32  `json:"pid"`
	Name       string `json:"name"`
	ExePath    string `json:"exe_path"`
	Cmdline    string `json:"cmdline"`
	CreateTime int64  `json:"create_time"`
}

type ETWNetworkEvent struct {
	PID        int32  `json:"pid"`
	Name       string `json:"name"`
	RemoteIP   string `json:"remote_ip"`
	RemotePort uint32 `json:"remote_port"`
	Proto      string `json:"proto"`
	Timestamp  int64  `json:"timestamp"`
}

type ETWRegistryEvent struct {
	PID       int32  `json:"pid"`
	Name      string `json:"name"`
	KeyPath   string `json:"key_path"`
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}

type ETWFileIOEvent struct {
	PID       int32  `json:"pid"`
	Name      string `json:"name"`
	FilePath  string `json:"file_path"`
	Operation string `json:"operation"`
	Timestamp int64  `json:"timestamp"`
}

func (s *ETWSensor) Name() string {
	return "etw"
}
