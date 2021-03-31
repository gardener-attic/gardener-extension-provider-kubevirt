// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package templates generated by go-bindata.// sources:
// scripts/download-cloud-config.tpl.sh
package templates

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _scriptsDownloadCloudConfigTplSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x92\x6f\x6b\xdb\x3c\x14\xc5\xdf\xeb\x53\xdc\xea\x31\x24\x79\xc0\x15\x85\xb1\x17\x1d\x1d\x18\xc5\x5b\x43\xd2\xa4\xd8\xe9\x46\x21\x2c\x28\xf2\x4d\x2d\xe2\xc8\x9e\xa4\x24\x0b\x6d\xbe\xfb\x70\x95\x3f\x2d\x84\xa4\xf3\xab\x0b\xba\xe7\x77\x2e\xc7\xe7\xbf\x0b\x36\x51\x9a\x4d\x84\xcd\x21\xc4\x05\x21\x69\xcc\x93\x78\x38\xee\x47\x77\xf1\x0d\x7d\x7e\x86\x4b\x8b\xd2\xa0\xeb\x8b\x39\xc2\x66\x43\x09\xb9\x8f\x86\xb7\x63\xde\x1b\x3c\xb4\xf9\xa0\xff\xad\xf3\x7d\xdc\x1e\xfc\xec\xf7\x06\x51\x3b\x4e\xc6\x69\x9c\xfc\x88\x13\xaf\xab\x84\xcb\xb9\xc1\x0c\xb5\x53\xa2\xb0\x29\x9a\x25\x9a\x57\xc4\x29\x02\x8f\xc6\x3c\x4e\x86\x47\x11\x3c\xe2\x68\xdc\x79\x44\xaf\x13\xf7\x87\x27\x30\x85\x42\xed\xfe\x05\xd5\x8d\x1f\x4f\x90\xba\xb8\x3e\x0e\xe2\xb7\x31\xef\xa6\x0f\x77\x07\x6d\xbb\x5c\xe9\xa2\x14\x19\x66\x3c\x47\x39\xb3\x8b\xb9\xcf\x54\x4d\xe1\x02\x7c\xf2\x37\x34\x68\xae\x9e\xd0\xc1\x88\x00\x84\xbf\x07\xa1\x1f\xc2\x1c\x45\x86\x06\x76\x1f\x8d\xa4\xc4\xca\x5d\x83\xa8\xaa\x42\x49\xe1\x54\xa9\xd9\x5a\xcc\x0b\xba\xdd\x97\x22\x94\x68\x9c\x9a\xd6\x8f\x08\x34\xf8\x40\xea\x7b\xed\x1b\x61\xed\x75\x5a\x7b\x88\x7b\xa7\xaf\x8c\x5a\x0a\x87\xe1\x0c\xd7\x1f\xd6\x77\xe3\x47\x2f\xa7\x41\x53\x0a\x77\x46\xe4\x8b\x46\x5b\x4c\x54\x8a\x2d\xaf\x98\x16\x73\xb4\x95\x90\x68\xd9\x6c\x31\xc1\xd0\xae\xad\xc3\x39\xf3\xdd\xb5\x2c\x78\xd3\x6a\xda\xa2\x5f\xc0\xe5\xa8\x09\x01\x40\x99\x97\x40\x79\xb9\x28\x32\xd0\xa5\x03\x83\xce\x28\x5c\x62\xbd\x00\xb2\x28\x17\x19\xc8\x52\x4f\xd5\x13\x58\x69\x54\xe5\x40\x69\xf0\x50\x58\x29\x97\x43\xed\x0b\xef\xe8\x35\xf4\x8f\x72\x70\x45\xa6\x8a\x90\x43\x05\x82\xa6\xf7\xda\x2e\x53\x78\x01\x8b\x19\x84\x46\x43\xc3\xb2\x3a\xa6\xba\x24\x42\xeb\xd2\xbd\xfe\xcc\x7d\x41\x5e\xc0\x60\x55\x08\x89\x40\x19\x05\x3a\x1a\x31\x0a\x9b\xcd\x35\x34\x2f\xff\x6f\xb1\xd1\x15\xab\x1a\x3b\x14\xd6\xa4\x5f\x94\xb1\xc6\x76\xa6\x01\x63\x8d\x16\x25\x5b\xe7\xdd\x31\x14\xbe\x1e\x8b\x77\xff\x4c\x48\xca\x93\xce\xfd\xf0\xfc\xd1\xf5\xc9\x99\x70\xa2\x8b\xeb\xd4\xe7\xf3\xfe\xb2\x83\xb7\x27\xd6\x80\x89\xb0\xf8\xf9\x13\x84\x99\x9f\x73\xf2\x37\x00\x00\xff\xff\x6a\x3c\x71\x3c\x83\x04\x00\x00")

func scriptsDownloadCloudConfigTplShBytes() ([]byte, error) {
	return bindataRead(
		_scriptsDownloadCloudConfigTplSh,
		"scripts/download-cloud-config.tpl.sh",
	)
}

func scriptsDownloadCloudConfigTplSh() (*asset, error) {
	bytes, err := scriptsDownloadCloudConfigTplShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "scripts/download-cloud-config.tpl.sh", size: 1155, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"scripts/download-cloud-config.tpl.sh": scriptsDownloadCloudConfigTplSh,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("nonexistent") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"scripts": &bintree{nil, map[string]*bintree{
		"download-cloud-config.tpl.sh": &bintree{scriptsDownloadCloudConfigTplSh, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
