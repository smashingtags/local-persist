package main

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

type localPersistDriver struct {
	volumes map[string]string
	mu      *sync.Mutex
	name    string
}

type saveData struct {
	State map[string]string `json:"state"`
}

func newLocalPersistDriver() localPersistDriver {
	driver := localPersistDriver{
		volumes: map[string]string{},
		mu:      &sync.Mutex{},
		name:    "local-persist",
	}

	os.MkdirAll(stateDir, 0700)

	if err, vols := loadStateFile(); err == nil {
		driver.volumes = vols
	}
	log.Printf("Starting with %d existing volumes", len(driver.volumes))

	return driver
}

func (d localPersistDriver) Get(req volume.Request) volume.Response {
	if mp, ok := d.volumes[req.Name]; ok {
		return volume.Response{Volume: &volume.Volume{Name: req.Name, Mountpoint: mp}}
	}
	return volume.Response{Err: fmt.Sprintf("no volume found with name %s", req.Name)}
}

func (d localPersistDriver) List(req volume.Request) volume.Response {
	var vols []*volume.Volume
	for name, mp := range d.volumes {
		vols = append(vols, &volume.Volume{Name: name, Mountpoint: mp})
	}
	return volume.Response{Volumes: vols}
}

func (d localPersistDriver) Create(req volume.Request) volume.Response {
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

func (d localPersistDriver) Remove(req volume.Request) volume.Response {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.volumes, req.Name)
	if err := d.saveState(); err != nil {
		log.Printf("warning: failed to save state: %v", err)
	}

	log.Printf("Removed volume %s", req.Name)
	return volume.Response{}
}

func (d localPersistDriver) Mount(req volume.MountRequest) volume.Response {
	return d.Path(volume.Request{Name: req.Name})
}

func (d localPersistDriver) Path(req volume.Request) volume.Response {
	return volume.Response{Mountpoint: d.volumes[req.Name]}
}

func (d localPersistDriver) Unmount(req volume.UnmountRequest) volume.Response {
	return d.Path(volume.Request{Name: req.Name})
}

func (d localPersistDriver) Capabilities(req volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func loadStateFile() (error, map[string]string) {
	p := filepath.Join(stateDir, stateFile)
	data, err := os.ReadFile(p)
	if err != nil {
		return err, nil
	}
	var s saveData
	if err := json.Unmarshal(data, &s); err != nil {
		return err, nil
	}
	return nil, s.State
}

func (d localPersistDriver) saveState() error {
	data, err := json.Marshal(saveData{State: d.volumes})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, stateFile), data, 0600)
}
