package service

type service interface {
	Close()
}

var services []service

func init() {
	services = []service{}
}

func AddService(s service) {
	services = append(services, s)
}

func ClosingServices() {
	for _, s := range services {
		if s != nil {
			s.Close()
		}
	}
}
