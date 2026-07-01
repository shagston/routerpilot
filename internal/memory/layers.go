package memory

import "fmt"

type Layer int

const (
	LayerWorking   Layer = 0
	LayerSession   Layer = 1
	LayerPersistent Layer = 2
	LayerExternal  Layer = 3
)

func (l Layer) String() string {
	switch l {
	case LayerWorking:
		return "working"
	case LayerSession:
		return "session"
	case LayerPersistent:
		return "persistent"
	case LayerExternal:
		return "external"
	default:
		return fmt.Sprintf("layer(%d)", int(l))
	}
}

var AllLayers = []Layer{LayerWorking, LayerSession, LayerPersistent, LayerExternal}
