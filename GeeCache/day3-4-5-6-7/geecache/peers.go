package geecache

import "GeeCache/day3-4-5-6-7/geecache/geecachepb"

//rpc client

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *geecachepb.Request, out *geecachepb.Response) error
}
