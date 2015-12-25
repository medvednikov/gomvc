package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		log.Fatal("usage: gomvc [command] [argument]")
	}
	command, arg := flag.Args()[0], flag.Args()[1]
	switch command {
	case "new":
		newProject(arg)
	default:
		log.Fatal("unknown command")
	}
}

func newProject(name string) {
	os.Mkdir(name, os.ModePerm)
	os.Mkdir(name+"/m", os.ModePerm)
	os.Mkdir(name+"/v", os.ModePerm)
	os.Mkdir(name+"/c", os.ModePerm)
	os.Mkdir(name+"/cmd", os.ModePerm)
	os.Mkdir(name+"/autogen", os.ModePerm)
	os.Mkdir(name+"/v/Home", os.ModePerm)

	newFile(name+"/cmd/main.go", `package main

import "github.com/medvednikov/gomvc"

import "../c"
import "../autogen"

func main() {
	gomvc.Route("/", &c.Home{})
	gomvc.Run(&gomvc.Config{
		Port:  "8080",
		IsDev: true,
		AssetNames: autogen.AssetNames(),
		AssetFunc:  autogen.Asset,
	})
}
`)

	newFile(name+"/c/home.go", `package c
import "github.com/medvednikov/gomvc"

type Home struct {
	*gomvc.Controller
}

func (c *Home) Index(name string) gomvc.View {
	return c.View(name)
}
`)

	newFile(name+"/v/Home/Index.html", `@t header

@if .
	Hello, @.!
@else
	Hello, world!
@end

@t footer
`)

	newFile(name+"/v/layout.html", `@define "header"
<!DOCTYPE html>
<html>
<head>
	<meta charset='utf-8'>
	<title>gomvc web applicaiton</title>
</head>
<body>
@end

@define "footer"
</body>
</html>
@end
`)

	newFile(name+"/autogen/autogen.go", `// This file has been generated automatically. Do not modify it.
package autogen

import "github.com/medvednikov/gomvc"

func init() {
	gomvc.ActionArgs = map[string]map[string][]string{"Home":map[string][]string{"Index":[]string{"name"}}}
}
`)

	newFile(name+"/autogen/templates.go", `
package autogen

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path"
	"path/filepath"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _home_index_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x72\x28\x51\xc8\x48\x4d\x4c\x49\x2d\xe2\xe2\x72\xc8\x4c\x53\xd0\xe3\xe2\xf4\x48\xcd\xc9\xc9\xd7\x51\x70\xd0\x53\xe4\x72\x48\xcd\x29\x4e\x85\x8b\x94\xe7\x17\xe5\xa4\x80\x04\xf3\x52\x80\x8a\x4b\x14\xd2\xf2\xf3\x4b\x80\xfa\x00\x01\x00\x00\xff\xff\x40\x8e\xe6\x88\x42\x00\x00\x00")

func home_index_html_bytes() ([]byte, error) {
	return bindata_read(
		_home_index_html,
		"Home/Index.html",
	)
}

func home_index_html() (*asset, error) {
	bytes, err := home_index_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "Home/Index.html", size: 66, mode: os.FileMode(420), modTime: time.Unix(1451051827, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _layout_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x3c\x8e\xbb\x0e\xc2\x30\x0c\x00\x67\xfc\x15\xa1\x4b\x27\x94\x95\xc1\x44\x95\x80\x19\x06\x16\xc6\x34\x71\x48\xa4\x3c\xaa\x62\x40\xfc\x3d\x25\x11\x4c\x96\x7c\x27\xfb\x06\x4b\x2e\x64\x12\x9d\x27\x6d\x69\xee\x00\xd7\x87\xd3\xfe\x72\x3d\x1f\x85\xe7\x14\x15\xe0\x6f\x2c\x5c\xc1\x0a\x13\xb1\x16\xc6\xeb\xf9\x4e\xbc\xeb\x1f\xec\x36\xdb\xfe\xbb\xe7\xc0\x91\xd4\xad\xa4\xa7\x11\x2f\x1a\x85\x9e\xa6\x18\x8c\x0e\x5c\x32\xca\x06\x01\x65\xbb\x82\x63\xb1\x6f\x05\x03\x65\x0b\xf0\x2f\x70\xa5\x70\x2d\x90\x0d\x2f\x76\x7d\x5d\xb5\x4f\x00\x00\x00\xff\xff\xaa\xdb\x60\x9e\xa8\x00\x00\x00")

func layout_html_bytes() ([]byte, error) {
	return bindata_read(
		_layout_html,
		"layout.html",
	)
}

func layout_html() (*asset, error) {
	bytes, err := layout_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "layout.html", size: 168, mode: os.FileMode(420), modTime: time.Unix(1451051827, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
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
	if (err != nil) {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
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
	"Home/Index.html": home_index_html,
	"layout.html": layout_html,
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
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
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
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() (*asset, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"Home": &_bintree_t{nil, map[string]*_bintree_t{
		"Index.html": &_bintree_t{home_index_html, map[string]*_bintree_t{
		}},
	}},
	"layout.html": &_bintree_t{layout_html, map[string]*_bintree_t{
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
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

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        if err != nil { // File
                return RestoreAsset(dir, name)
        } else { // Dir
                for _, child := range children {
                        err = RestoreAssets(dir, path.Join(name, child))
                        if err != nil {
                                return err
                        }
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
`)

}

func newFile(name, text string) {
	file, err := os.Create(name)
	if err != nil {
		log.Fatal("failed to create file", name)
	}
	fmt.Fprint(file, text)
	file.Close()
}
