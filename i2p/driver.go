package ovs

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/client"
	"github.com/eyedeekay/onramp"

	dknet "github.com/docker/go-plugins-helpers/network"
	log "github.com/sirupsen/logrus"
)

const (
	defaultRoute     = "0.0.0.0/0"
	ovsPortPrefix    = "ovs-veth0-"
	bridgePrefix     = "ovsbr-"
	containerEthName = "eth"

	mtuOption           = "net.gopher.ovs.bridge.mtu"
	modeOption          = "net.gopher.ovs.bridge.mode"
	bridgeNameOption    = "net.gopher.ovs.bridge.name"
	bindInterfaceOption = "net.gopher.ovs.bridge.bind_interface"

	modeNAT  = "nat"
	modeFlat = "flat"

	defaultMTU  = 1500
	defaultMode = modeNAT
)

var (
	validModes = map[string]bool{
		modeNAT:  true,
		modeFlat: true,
	}
)

type Driver struct {
	parent dknet.Driver
	dockerer
	//ovsdber
	networks map[string]*NetworkState
	samConn  *onramp.Garlic
	//OvsdbNotifier
}

// NetworkState is filled in at network creation time
// it contains state that we wish to keep for each network
type NetworkState struct {
	BridgeName        string
	MTU               int
	Mode              string
	Gateway           string
	GatewayMask       string
	FlatBindInterface string
}

func (d *Driver) ProgramExternalConnectivity(r *dknet.ProgramExternalConnectivityRequest) error {
	return nil
}

func (d *Driver) RevokeExternalConnectivity(r *dknet.RevokeExternalConnectivityRequest) error {
	return nil
}

func (d *Driver) GetCapabilities() (*dknet.CapabilitiesResponse, error) {
	return nil, nil
}

func (d *Driver) FreeNetwork(r *dknet.FreeNetworkRequest) error {
	return nil
}

func (d *Driver) DiscoverNew(r *dknet.DiscoveryNotification) error {
	return nil
}
func (d *Driver) AllocateNetwork(r *dknet.AllocateNetworkRequest) (*dknet.AllocateNetworkResponse, error) {
	log.Debugf("Allocate network request: %+v", r)
	return nil, nil
}

func (d *Driver) DiscoverDelete(r *dknet.DiscoveryNotification) error {
	return nil
}

func (d *Driver) CreateNetwork(r *dknet.CreateNetworkRequest) error {
	log.Debugf("Create network request: %+v", r)

	bridgeName, err := getBridgeName(r)
	if err != nil {
		return err
	}

	mtu, err := getBridgeMTU(r)
	if err != nil {
		return err
	}

	mode, err := getBridgeMode(r)
	if err != nil {
		return err
	}

	gateway, mask, err := getGatewayIP(r)
	if err != nil {
		return err
	}

	bindInterface, err := getBindInterface(r)
	if err != nil {
		return err
	}

	ns := &NetworkState{
		BridgeName:        bridgeName,
		MTU:               mtu,
		Mode:              mode,
		Gateway:           gateway,
		GatewayMask:       mask,
		FlatBindInterface: bindInterface,
	}
	d.networks[r.NetworkID] = ns

	log.Debugf("Initializing bridge for network %s", r.NetworkID)
	//if err := d.initBridge(r.NetworkID); err != nil {
	//delete(d.networks, r.NetworkID)
	//return err
	//}
	return nil
}

func (d *Driver) DeleteNetwork(r *dknet.DeleteNetworkRequest) error {
	log.Debugf("Delete network request: %+v", r)
	delete(d.networks, r.NetworkID)
	return nil
}

func (d *Driver) CreateEndpoint(r *dknet.CreateEndpointRequest) (*dknet.CreateEndpointResponse, error) {
	log.Debugf("Create endpoint request: %+v", r)
	return nil, nil
}

func (d *Driver) DeleteEndpoint(r *dknet.DeleteEndpointRequest) error {
	log.Debugf("Delete endpoint request: %+v", r)
	return nil
}

func (d *Driver) EndpointInfo(r *dknet.InfoRequest) (*dknet.InfoResponse, error) {
	res := &dknet.InfoResponse{
		Value: make(map[string]string),
	}
	return res, nil
}

func (d *Driver) Join(r *dknet.JoinRequest) (*dknet.JoinResponse, error) {
	res := &dknet.JoinResponse{
		InterfaceName: dknet.InterfaceName{
			SrcName:   d.samConn.B32(),
			DstPrefix: containerEthName,
		},
		Gateway: d.networks[r.NetworkID].Gateway,
	}
	log.Debugf("Join endpoint %s:%s to %s", r.NetworkID, r.EndpointID, r.SandboxKey)
	return res, nil
}

func (d *Driver) Leave(r *dknet.LeaveRequest) error {
	log.Debugf("Leave request: %+v", r)
	log.Debugf("Leave %s:%s", r.NetworkID, r.EndpointID)
	return nil
}

func NewDriver() (*Driver, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %s", err)
	}

	// initiate the tunnel manager port binding
	var samConn *onramp.Garlic
	retries := 3
	for i := 0; i < retries; i++ {
		samConn, err = onramp.NewGarlic("i2p-driver", "127.0.0.1:7656", onramp.OPT_DEFAULTS)
		if err == nil {
			break
		}
		log.Errorf("could not connect to I2P on port [ %d ]: %s. Retrying in 5 seconds", "127.0.0.1:7656", err)
		time.Sleep(5 * time.Second)
	}

	if samConn == nil {
		return nil, fmt.Errorf("could not connect to I2P")
	}

	d := &Driver{
		dockerer: dockerer{
			client: docker,
		},
		networks: make(map[string]*NetworkState),
	}
	return d, nil
}

func truncateID(id string) string {
	return id[:5]
}

func getBridgeMTU(r *dknet.CreateNetworkRequest) (int, error) {
	bridgeMTU := defaultMTU
	if r.Options != nil {
		if mtu, ok := r.Options[mtuOption].(int); ok {
			bridgeMTU = mtu
		}
	}
	return bridgeMTU, nil
}

func getBridgeName(r *dknet.CreateNetworkRequest) (string, error) {
	bridgeName := bridgePrefix + truncateID(r.NetworkID)
	if r.Options != nil {
		if name, ok := r.Options[bridgeNameOption].(string); ok {
			bridgeName = name
		}
	}
	return bridgeName, nil
}

func getBridgeMode(r *dknet.CreateNetworkRequest) (string, error) {
	return "overlay", nil
}

func getGatewayIP(r *dknet.CreateNetworkRequest) (string, string, error) {
	// FIXME: Dear future self, I'm sorry for leaving you with this mess, but I want to get this working ASAP
	// This should be an array
	// We need to handle case where we have
	// a. v6 and v4 - dual stack
	// auxilliary address
	// multiple subnets on one network
	// also in that case, we'll need a function to determine the correct default gateway based on it's IP/Mask
	var gatewayIP string
	if len(r.IPv6Data) > 0 {
		if r.IPv6Data[0] != nil {
			if r.IPv6Data[0].Gateway != "" {
				gatewayIP = r.IPv6Data[0].Gateway
			}
		}
	}
	// Assumption: IPAM will provide either IPv4 OR IPv6 but not both
	// We may want to modify this in future to support dual stack
	if len(r.IPv4Data) > 0 {
		if r.IPv4Data[0] != nil {
			if r.IPv4Data[0].Gateway != "" {
				gatewayIP = r.IPv4Data[0].Gateway
			}
		}
	}
	if gatewayIP == "" {
		return "", "", fmt.Errorf("No gateway IP found")
	}
	parts := strings.Split(gatewayIP, "/")
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Cannot split gateway IP address")
	}
	return parts[0], parts[1], nil
}

func getBindInterface(r *dknet.CreateNetworkRequest) (string, error) {
	if r.Options != nil {
		if mode, ok := r.Options[bindInterfaceOption].(string); ok {
			return mode, nil
		}
	}
	// As bind interface is optional and has no default, don't return an error
	return "", nil
}
