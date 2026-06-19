package collector

type Sensor interface {
	Name() string
	Collect() (interface{}, error)
}

var Registry []Sensor

func Register(s Sensor) {
	Registry = append(Registry, s)
}

func InitCollectors() {
	Register(&SoftwareSensor{})
	Register(&USBSensor{})
	Register(&PortSensor{})
	Register(&InventorySensor{})
	Register(&AntivirusSensor{})
	Register(&DataTransferSensor{})
	Register(&ETWSensor{
		Filter: EventFilter{
			EnableProcess:  true,
			EnableNetwork:  true,
			EnableRegistry: true,
			EnableFileIO:   true,
		},
	})
	Register(&PatchSensor{})
}
