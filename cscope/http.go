package cscope

//go:generate go-bindata -pkg $GOPACKAGE -o assets.go -prefix "../assets/" ../assets/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"html/template"
	"net/http"

	"github.com/edwardchoh/jade"
	"github.com/go-playground/log"
)

type HttpServer struct {
	*http.Server

	cscope       *Cscope
	assetsPath   string
	templates    map[string]*template.Template
	useLiveFiles bool
}

// type AssetLoader func(string) ([]byte, error)

func NewHttpServer(listenAddr, pathToCscopeOut, assetsPath string) (h *HttpServer, err error) {
	s := &http.Server{
		Addr:           listenAddr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	var cs *Cscope
	if cs, err = NewCscope("cscope", pathToCscopeOut); err != nil {
		return
	}

	assets := []string{}
	for _, name := range AssetNames() {
		if path.Ext(name) == ".jade" {
			assets = append(assets, name)
		}
	}

	var templates map[string]*template.Template
	templates, err = jade.CompileAssets(assets, Asset, jade.Options{})
	if err != nil {
		return
	}

	h = &HttpServer{s, cs, assetsPath, templates, assetsPath != ""}
	s.Handler = h

	return
}

func (h *HttpServer) getTemplate(template string) *template.Template {
	if h.useLiveFiles {
		tpl, err := jade.CompileFile(h.getResourcePath(template+".jade"), jade.Options{})
		if err != nil {
			panic(err)
		}
		return tpl
	} else {
		return h.templates[template]
	}
}

func (h *HttpServer) getResourcePath(filename string) string {
	return path.Join(h.assetsPath, filename)
}

func (h *HttpServer) ServeAsset(w http.ResponseWriter, req *http.Request) {
	var buf []byte
	var err error

	file := req.URL.Query()["file"][0]

	if h.useLiveFiles {
		buf, err = ioutil.ReadFile(h.getResourcePath(file))
	} else {
		// serve file from assets
		buf, err = Asset(file)
	}

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		ext := path.Ext(file)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}
		w.Header().Set("Content-Type", "text/"+ext)
		if !h.useLiveFiles {
			w.Header().Set("Cache-Control", "max-age=120")
		}
		w.Write(buf)
	}
	return
}

func (h *HttpServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Path == "/" && len(req.URL.Query()) != 0 {
		if _, ok := req.URL.Query()["file"]; ok {
			h.ServeAsset(w, req)
		} else {
			h.ServeQuery(w, req)
		}
		return
	}

	p := path.Join(h.cscope.SourcePath, req.URL.Path)
	f, err := os.Open(p)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if fi.Mode().IsDir() {
		h.ServeDir(w, req)
	} else if fi.Mode().IsRegular() {
		h.ServeFile(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	return
}

func (h *HttpServer) ServeQuery(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{}

	if req.URL.Query()["kind"] != nil && req.URL.Query()["search"] != nil {
		k := req.URL.Query()["kind"][0]
		v := req.URL.Query()["search"][0]

		if kind, ok := CscopeKind[strings.ToLower(k)]; ok {
			res, err := h.cscope.Search(v, kind)

			if err != nil {
				log.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// group res by dir name/file name
			r := make(map[string]map[string][]*CscopeResult)
			for _, v := range res {
				k := path.Dir(v.File)
				p := path.Base(v.File)

				if len(k) > 0 && k[0] == '/' {
					// ignore results outside of the cscope root
					continue
				}

				f, ok := r[k]
				if !ok {
					f = make(map[string][]*CscopeResult)
					r[k] = f
				}

				rs, ok := f[p]
				if !ok {
					rs = make([]*CscopeResult, 0)
					f[p] = rs
				}
				f[p] = append(rs, v)
			}

			data["Results"] = r
			data["Query"] = v
		}
	}

	tpl := h.getTemplate("query")
	err := tpl.Execute(w, data)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HttpServer) ServeDir(w http.ResponseWriter, req *http.Request) {
	p := path.Clean(path.Join(h.cscope.SourcePath, req.URL.Path))
	fis, err := ioutil.ReadDir(p)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	files := map[string]map[string]string{}
	data := map[string]interface{}{
		"Files": files,
		"Path":  req.URL.Path,
	}

	// if req.URL.Path != "/" {
	// 	files["../"] = "../"
	// }

	for _, fi := range fis {
		if fi.IsDir() {
			f := fi.Name() + "/"
			files[fi.Name()+"/"] = map[string]string{
				"url":  path.Join(req.URL.Path, f),
				"date": fi.ModTime().Format("02 Jan 06 15:04"),
				"size": "",
			}
		} else {
			f := fi.Name()
			files[fi.Name()] = map[string]string{
				"url":  path.Join(req.URL.Path, f),
				"date": fi.ModTime().Format("02 Jan 06 15:04"),
				"size": fmt.Sprintf("%d", fi.Size()),
			}
		}
	}

	tpl := h.getTemplate("directory")
	err = tpl.Execute(w, data)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (h *HttpServer) ServeFile(w http.ResponseWriter, req *http.Request) {
	p := path.Join(h.cscope.SourcePath, req.URL.Path)
	content, err := ioutil.ReadFile(p)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	paths := []map[string]string{}
	for pt := path.Clean(req.URL.Path); pt != "/"; pt = path.Dir(pt) {
		b := path.Base(pt)
		paths = append([]map[string]string{map[string]string{
			"name": b,
			"path": pt,
		}}, paths...)
	}
	paths = append([]map[string]string{map[string]string{
		"name": "/",
		"path": "/",
	}}, paths...)

	data := map[string]interface{}{
		"Content": string(content),
		"Path":    req.URL.Path,
		"Paths":   paths,
	}

	tpl := h.getTemplate("file")

	err = tpl.Execute(w, data)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}
