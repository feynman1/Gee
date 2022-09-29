package geecache

//rpc server
import (
	consistenthhash "GeeCache/day3-4-5-6-7/geecache/consistenthash"
	"GeeCache/day3-4-5-6-7/geecache/geecachepb"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *geecachepb.Request, out *geecachepb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	res, errg := http.Get(u)
	if errg != nil {
		return errg
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server return:%v", res.Status)
	}

	bytes, erri := ioutil.ReadAll(res.Body)
	if erri != nil {
		return fmt.Errorf("reading response body: ", erri)
	}

	errp := proto.Unmarshal(bytes, out)
	if errp != nil {
		return fmt.Errorf("decoding response body: %v", errp)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)

type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self       string
	basePath   string
	mu         sync.Mutex //guards peers and httpGetters
	peers      *consistenthhash.Map
	httpGetter map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthhash.New(nil, defaultReplicas)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetter[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path :" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}
	v, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err1 := proto.Marshal(&geecachepb.Response{Value: v.ByteSlice()})
	if err1 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Context-Type", "application/octet-stream")
	w.Write(body)
}
