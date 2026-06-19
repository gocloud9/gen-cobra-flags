package subcommands

// CreatePeer is a child resource referenced by CreateNetworkRequest. It is not
// directly enabled; it is pulled into generation via the +cobra:config:child /
// +cobra:subcommand markers on the parent field. Its +cobra:required fields are
// auto-derived into individual flags.
type CreatePeer struct {
	// +cobra:required
	PeerVpcId string
	// +cobra:required
	PeerRegion string
	Note       string
}

// CreateNetworkRequest is the enabled top-level request.
// +cobra:enabled
// +cobra:flag=config
// +cobra:short=c
// +cobra:usage=Aggregate config for the network request.
// +cobra:subcommand:config:prefix=Network
type CreateNetworkRequest struct {
	// +cobra:flag=name
	// +cobra:short=n
	// +cobra:usage=Network name.
	// +cobra:required
	Name string

	// +cobra:flag=peers
	// +cobra:usage=Peers to add to the network.
	// +cobra:config:child
	// +cobra:subcommand
	// +cobra:subcommand:config:flag=peer-config
	// +cobra:subcommand:config:short=p
	// +cobra:subcommand:config:usage=The configuration for a network peer to be added.
	Peers []*CreatePeer

	// RemovePeers is a scalar-slice subcommand: a repeated primitive field with
	// no struct child. The subcommand surfaces a single value flag whose value
	// is appended to this slice.
	// +cobra:flag=remove-peers
	// +cobra:usage=Peers to remove from the network.
	// +cobra:subcommand
	// +cobra:subcommand:config:flag=remove-peer-config
	// +cobra:subcommand:config:usage=The configuration for a network peer to be removed.
	// +cobra:subcommand:value:flag=remove-peer-vpc-id
	// +cobra:subcommand:value:short=r
	// +cobra:subcommand:value:usage=The vpc id of the peer to remove.
	RemovePeers []string
}
