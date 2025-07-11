// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package types

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/cilium/cilium/pkg/identity"
)

// IdentityUpdater is responsible for handling identity updates into the core
// policy engine. See SelectorCache.UpdateIdentities() for more details.
type IdentityUpdater interface {
	UpdateIdentities(added, deleted identity.IdentityMap) <-chan struct{}
}

// ResourceID identifies a unique copy of a resource that provides a source for
// information tied to an IP address in the IPCache.
type ResourceID string

// ResourceKind determines the source of the ResourceID. Typically this is the
// short name for the k8s resource.
type ResourceKind string

var (
	ResourceKindCCNP      = ResourceKind("ccnp")
	ResourceKindCIDRGroup = ResourceKind("cidrgroup")
	ResourceKindCNP       = ResourceKind("cnp")
	ResourceKindDaemon    = ResourceKind("daemon")
	ResourceKindEndpoint  = ResourceKind("ep")
	ResourceKindFile      = ResourceKind("file")
	ResourceKindNetpol    = ResourceKind("netpol")
	ResourceKindNode      = ResourceKind("node")
)

// NewResourceID returns a ResourceID populated with the standard fields for
// uniquely identifying a source of IPCache information.
func NewResourceID(kind ResourceKind, namespace, name string) ResourceID {
	str := strings.Builder{}
	str.Grow(len(kind) + 1 + len(namespace) + 1 + len(name))
	str.WriteString(string(kind))
	str.WriteRune('/')
	str.WriteString(namespace)
	str.WriteRune('/')
	str.WriteString(name)
	return ResourceID(str.String())
}

func (r ResourceID) Namespace() string {
	parts := strings.SplitN(string(r), "/", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// TunnelPeer is the IP address of the host associated with this prefix. This is
// typically used to establish a tunnel, e.g. in tunnel mode or for encryption.
// This type implements ipcache.IPMetadata
type TunnelPeer struct{ netip.Addr }

func (t TunnelPeer) IP() net.IP {
	return t.AsSlice()
}

// EncryptKey is the identity of the encryption key.
// This type implements ipcache.IPMetadata
type EncryptKey uint8

const EncryptKeyEmpty = EncryptKey(0)

func (e EncryptKey) IsValid() bool {
	return e != EncryptKeyEmpty
}

func (e EncryptKey) Uint8() uint8 {
	return uint8(e)
}

func (e EncryptKey) String() string {
	return strconv.Itoa(int(e))
}

// RequestedIdentity is a desired numeric identity for the prefix. When the
// prefix is next injected, this numeric ID will be requested from the local
// allocator. If the allocator can accommodate that request, it will do so.
// In order for this to be useful, the prefix must not already have an identity
// (or its set of labels must have changed), and that numeric identity must
// be free.
// Thus, the numeric ID should have already been held-aside in the allocator
// using WithholdLocalIdentities(). That will ensure the numeric ID remains free
// for the prefix to request.
type RequestedIdentity identity.NumericIdentity

func (id RequestedIdentity) IsValid() bool {
	return id != 0
}

func (id RequestedIdentity) ID() identity.NumericIdentity {
	return identity.NumericIdentity(id)
}

// EndpointFlags represents various flags that can be attached to endpoints in the IPCache
// This type implements ipcache.IPMetadata
type EndpointFlags struct {
	// isInit gets flipped to true on the first intentional flag set
	// it is a sentinel to distinguish an uninitialized EndpointFlags
	// from one with all flags set to false
	isInit bool

	// flagSkipTunnel can be applied to a remote endpoint to signal that
	// packets destined for said endpoint shall not be forwarded through
	// an overlay tunnel, regardless of Cilium's configuration.
	flagSkipTunnel bool
}

func (e *EndpointFlags) SetSkipTunnel(skip bool) {
	e.isInit = true
	e.flagSkipTunnel = skip
}

func (e EndpointFlags) IsValid() bool {
	return e.isInit
}

// Uint8 encoding MUST mimic the one in pkg/maps/ipcache
// since it will eventually get recast to it
const (
	FlagSkipTunnel uint8 = 1 << iota
)

func (e EndpointFlags) Uint8() uint8 {
	var flags uint8 = 0
	if e.flagSkipTunnel {
		flags = flags | FlagSkipTunnel
	}
	return flags
}
