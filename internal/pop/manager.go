package pop

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"titan-ipweb/internal/types"
	"titan-ipweb/ippmclient"

	"golang.org/x/sync/singleflight"
)

// Resolve the pop update
type Manager struct {
	ippmServerURL string
	accessToken   string
	mu            sync.RWMutex
	pops          map[string]*types.Pop
	group         singleflight.Group
}

func NewPopManager(ippmServerURL, accessToken string) (*Manager, error) {
	m := &Manager{
		ippmServerURL: ippmServerURL,
		accessToken:   accessToken,
		mu:            sync.RWMutex{},
		pops:          make(map[string]*types.Pop),
		group:         singleflight.Group{},
	}
	pops, err := m.fetch(ippmServerURL, accessToken)
	if err != nil {
		return nil, err
	}
	m.pops = pops
	return m, nil
}

func (m *Manager) Get(popID string) (*types.Pop, error) {
	m.mu.RLock()
	if p, ok := m.pops[popID]; ok {
		m.mu.RUnlock()
		return p, nil
	}
	m.mu.RUnlock()

	v, err, _ := m.group.Do("fetch_pops", func() (interface{}, error) {
		pops, err := m.fetch(m.ippmServerURL, m.accessToken)
		if err != nil {
			return nil, err
		}

		m.mu.Lock()
		m.pops = pops
		m.mu.Unlock()

		return pops, nil
	})
	if err != nil {
		return nil, err
	}

	pops := v.(map[string]*types.Pop)
	if p, ok := pops[popID]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("pop %s not exist", popID)
}

func (m *Manager) fetch(ippmServer, accessToken string) (map[string]*types.Pop, error) {
	url := fmt.Sprintf("%s/pops", ippmServer)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http status code %d, error msg %s", resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	popsResp := &ippmclient.GetPopsResp{}
	err = json.Unmarshal(data, popsResp)
	if err != nil {
		return nil, err
	}

	// pops := make([]*types.Pop, 0, len(popsResp.Pops))
	pops := make(map[string]*types.Pop)
	for _, p := range popsResp.Pops {
		pop := &types.Pop{ID: p.ID, Name: p.Name, Area: p.Area, Socks5Server: p.Socks5Addr}
		pops[pop.ID] = pop
	}
	return pops, nil
}
