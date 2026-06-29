package network

type MockProvider struct {
	GetInterfaceStatusFunc func() ([]InterfaceStatus, error)
	SetInterfaceStateFunc func(name string, up bool) error
	GetAddressesFunc      func() ([]Address, error)
	AddAddressFunc        func(iface, addr string) error
	GetRoutesFunc         func() ([]Route, error)
	AddRouteFunc          func(dest, gateway, iface string) error
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		GetInterfaceStatusFunc: func() ([]InterfaceStatus, error) {
			return []InterfaceStatus{
				{Name: "lo", Up: true, Active: true},
				{Name: "eth0", Up: true, Active: true},
				{Name: "wlan0", Up: true, Active: true},
			}, nil
		},
		SetInterfaceStateFunc: func(name string, up bool) error {
			return nil
		},
		GetAddressesFunc: func() ([]Address, error) {
			return []Address{
				{Interface: "lo", Address: "127.0.0.1/8"},
				{Interface: "eth0", Address: "192.168.1.100/24"},
			}, nil
		},
		AddAddressFunc: func(iface, addr string) error {
			return nil
		},
		GetRoutesFunc: func() ([]Route, error) {
			return []Route{
				{Destination: "default", Gateway: "192.168.1.1", Interface: "eth0"},
				{Destination: "192.168.1.0/24", Gateway: "", Interface: "eth0"},
			}, nil
		},
		AddRouteFunc: func(dest, gateway, iface string) error {
			return nil
		},
	}
}

func (m *MockProvider) GetInterfaceStatus() ([]InterfaceStatus, error) {
	if m.GetInterfaceStatusFunc != nil {
		return m.GetInterfaceStatusFunc()
	}
	return nil, nil
}

func (m *MockProvider) SetInterfaceState(name string, up bool) error {
	if m.SetInterfaceStateFunc != nil {
		return m.SetInterfaceStateFunc(name, up)
	}
	return nil
}

func (m *MockProvider) GetAddresses() ([]Address, error) {
	if m.GetAddressesFunc != nil {
		return m.GetAddressesFunc()
	}
	return nil, nil
}

func (m *MockProvider) AddAddress(iface, addr string) error {
	if m.AddAddressFunc != nil {
		return m.AddAddressFunc(iface, addr)
	}
	return nil
}

func (m *MockProvider) GetRoutes() ([]Route, error) {
	if m.GetRoutesFunc != nil {
		return m.GetRoutesFunc()
	}
	return nil, nil
}

func (m *MockProvider) AddRoute(dest, gateway, iface string) error {
	if m.AddRouteFunc != nil {
		return m.AddRouteFunc(dest, gateway, iface)
	}
	return nil
}
