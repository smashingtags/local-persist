package driver

import (
	"os"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	testName       = "test-volume"
	testMountpoint = "/tmp/data/local-persist-test"
)

func createVolume(t *testing.T, d Driver, name, mountpoint string) {
	t.Helper()
	res := d.Create(volume.Request{
		Name:    name,
		Options: map[string]string{"mountpoint": mountpoint},
	})
	if res.Err != "" {
		t.Fatal(res.Err)
	}
}

func removeVolume(t *testing.T, d Driver, name, mountpoint string) {
	t.Helper()
	os.RemoveAll(mountpoint)
	d.Remove(volume.Request{Name: name})
	res := d.Get(volume.Request{Name: name})
	if res.Err == "" {
		t.Error("volume should not exist after removal")
	}
}

func TestCreate(t *testing.T) {
	d := New()
	createVolume(t, d, testName, testMountpoint)

	if _, err := os.Stat(testMountpoint); os.IsNotExist(err) {
		t.Error("mountpoint directory was not created")
	}
	if len(d.volumes) != 1 {
		t.Errorf("expected 1 volume, got %d", len(d.volumes))
	}

	removeVolume(t, d, testName, testMountpoint)

	res := d.Create(volume.Request{Name: testName})
	if res.Err != "the `mountpoint` option is required" {
		t.Error("should require mountpoint option")
	}
}

func TestGet(t *testing.T) {
	d := New()
	createVolume(t, d, testName, testMountpoint)

	res := d.Get(volume.Request{Name: testName})
	if res.Err != "" {
		t.Error("should find volume")
	}

	removeVolume(t, d, testName, testMountpoint)
}

func TestList(t *testing.T) {
	d := New()
	createVolume(t, d, testName, testMountpoint)

	res := d.List(volume.Request{})
	if len(res.Volumes) != 1 {
		t.Errorf("expected 1 volume, got %d", len(res.Volumes))
	}

	createVolume(t, d, testName+"2", testMountpoint+"2")
	res = d.List(volume.Request{})
	if len(res.Volumes) != 2 {
		t.Errorf("expected 2 volumes, got %d", len(res.Volumes))
	}

	removeVolume(t, d, testName, testMountpoint)
	removeVolume(t, d, testName+"2", testMountpoint+"2")
}

func TestMountUnmountPath(t *testing.T) {
	d := New()
	createVolume(t, d, testName, testMountpoint)

	pathRes := d.Path(volume.Request{Name: testName})
	mountRes := d.Mount(volume.MountRequest{Name: testName})
	unmountRes := d.Unmount(volume.UnmountRequest{Name: testName})

	if pathRes.Mountpoint != testMountpoint ||
		mountRes.Mountpoint != testMountpoint ||
		unmountRes.Mountpoint != testMountpoint {
		t.Error("Mount, Unmount, and Path should all return the same mountpoint")
	}
}
