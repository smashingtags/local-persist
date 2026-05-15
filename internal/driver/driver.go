package driver

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	stateDir  = "/var/lib/docker/plugin-data/"
	stateFile = "local-persist.json"
)

type Driver struct {
	volumes map[string]string
	mu      *sync.Mutex
}

type saveData struct {
	State map[string]string `json:"state"`
}

func New() Driver {
	d := Driver{
		volumes: map[string]string{},
		mu:      &sync.Mutex{},
	}

	os.MkdirAll(stateDir, 0700)

	if vols, err := loadStateFile(); err == nil {
		d.volumes = vols
	}
	log.Printf("Starting with %d existing volumes", len(d.volumes))

	return d
}

func (d Driver) Get(req volume.Request) volume.Response {
	if mp, ok := d.volumes[req.Name]; ok {
		return volume.Response{Volume: &volume.Volume{Name: req.Name, Mountpoint: mp}}
	}
	return volume.Response{Err: fmt.Sprintf("no volume found with name %s", req.Name)}
}

func (d Driver) List(req volume.Request) volume.Response {
	var vols []*volume.Volume
	for name, mp := range d.volumes {
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: mp})
	}
	return volume.Response{Volumes: vols}
}

func (d Driver) Create(req volume.Request) volume.Response {
	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		return volume.Response{Err: "the `mountpoint` option is required"}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.volumes[req.Name]; exists {
		return volume.Response{Err: fmt.Sprintf("volume %s already exists", req.Name)}
	}

	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		return volume.Response{Err: fmt.Sprintf("could not create directory %s: %v", mountpoint, err)}
	}

	d.volumes[req.Name] = mountpoint
	if err := d.saveState(); err != nil {
		log.Printf("warning: failed to save state: %v", err)
	}

	log.Printf("Created volume %s at %s", req.Name, mountpoint)
	return volume.Response{}
}

func (d Driver) Remove(req volume.Request) volume.Response {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.volumes, req.Name)
	if err := d.saveState(); err != nil {
		log.Printf("warning: failed to save state: %v", err)
	}

	log.Printf("Removed volume %s", req.Name)
	return volume.Response{}
}

func (d Driver) Mount(req volume.MountRequest) volume.Response {
	return d.Path(volume.Request{Name: req.Name})
}

func (d Driver) Path(req volume.Request) volume.Response {
	return volume.Response{Mountpoint: d.volumes[req.Name]}
}

func (d Driver) Unmount(req volume.UnmountRequest) volume.Response {
	return d.Path(volume.Request{Name: req.Name})
}

func (d Driver) Capabilities(req volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func loadStateFile() (map[string]string, error) {
	p := filepath.Join(stateDir, stateFile)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var s saveData
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return s.State, nil
}

func (d Driver) saveState() error {
	data, err := json.Marshal(saveData{State: d.volumes})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, stateFile), data, 0600)
}
