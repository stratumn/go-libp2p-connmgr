package connmgr

import (
	"context"
	"testing"

	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	tu "github.com/libp2p/go-testutil"
)

type tconn struct {
	inet.Conn
	peer   peer.ID
	closed bool
}

func (c *tconn) Close() error {
	c.closed = true
	return nil
}

func (c *tconn) RemotePeer() peer.ID {
	return c.peer
}

func randConn(t *testing.T) inet.Conn {
	pid := tu.RandPeerIDFatal(t)
	return &tconn{peer: pid}
}

func TestConnTrimming(t *testing.T) {
	cm := NewConnManager(200, 300, 0)
	not := cm.Notifee()

	var conns []inet.Conn
	for i := 0; i < 300; i++ {
		rc := randConn(t)
		conns = append(conns, rc)
		not.Connected(nil, rc)
	}

	for _, c := range conns {
		if c.(*tconn).closed {
			t.Fatal("nothing should be closed yet")
		}
	}

	for i := 0; i < 100; i++ {
		cm.TagPeer(conns[i].RemotePeer(), "foo", 10)
	}

	cm.TagPeer(conns[299].RemotePeer(), "badfoo", -5)

	cm.TrimOpenConns(context.Background())

	for i := 0; i < 100; i++ {
		c := conns[i]
		if c.(*tconn).closed {
			t.Fatal("these shouldnt be closed")
		}
	}

	if !conns[299].(*tconn).closed {
		t.Fatal("conn with bad tag should have gotten closed")
	}
}
