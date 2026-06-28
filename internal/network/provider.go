package network

type InterfaceStatus struct {
	Name   string `json:"name"`
	Up     bool   `json:"up"`
	Active bool   `json:"active"`
}

type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
}

type Address struct {
	Interface string `json:"interface"`
	Address   string `json:"address"`
}

type Provider interface {
	GetInterfaceStatus() ([]InterfaceStatus, error)
	SetInterfaceState(name string, up bool) error

	GetAddresses() ([]Address, error)
	AddAddress(iface, addr string) error
	GetRoutes() ([]Route, error)
	AddRoute(dest, gateway, iface string) error
}
