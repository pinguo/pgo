package pgo

import (
    "bytes"
    "fmt"
    "html/template"
    "io"
    "path/filepath"
    "sync"
)

// View the view component, configuration:
// view:
//     suffix: ".html"
//     commons:
//         - "@view/common/header.html"
//         - "@view/common/footer.html"
type View struct {
    suffix    string
    commons   []string
    funcMap   template.FuncMap
    templates map[string]*template.Template
    lock      sync.RWMutex
}

func (v *View) Construct() {
    v.suffix = ".html"
    v.commons = make([]string, 0)
    v.templates = make(map[string]*template.Template)
}

// SetSuffix set view file suffix, default is ".html"
func (v *View) SetSuffix(suffix string) {
    if len(suffix) > 0 && suffix[0] != '.' {
        suffix = "." + suffix
    }

    if len(suffix) > 1 {
        v.suffix = suffix
    }
}

// SetCommons set common view files
func (v *View) SetCommons(commons []interface{}) {
    for _, p := range commons {
        if view, ok := p.(string); ok {
            v.commons = append(v.commons, v.normalize(view))
        } else {
            panic(fmt.Sprintf("invalid common view, %s", p))
        }
    }
}

// AddFuncMap add custom func map
func (v *View) AddFuncMap(funcMap template.FuncMap) {
    v.funcMap = funcMap
}

// Render render view and return result
func (v *View) Render(view string, data interface{}) []byte {
    buf := &bytes.Buffer{}
    v.Display(buf, view, data)
    return buf.Bytes()
}

// Display render view and display result
func (v *View) Display(w io.Writer, view string, data interface{}) {
    view = v.normalize(view)
    tpl := v.getTemplate(view)
    e := tpl.Execute(w, data)
    if e != nil {
        panic(fmt.Sprintf("failed to render view, %s, %s", view, e))
    }
}

func (v *View) getTemplate(view string) *template.Template {
    if _, ok := v.templates[view]; !ok {
        v.loadTemplate(view)
    }

    v.lock.RLock()
    defer v.lock.RUnlock()

    return v.templates[view]
}

func (v *View) loadTemplate(view string) {
    v.lock.Lock()
    defer v.lock.Unlock()

    // avoid repeated loading
    if _, ok := v.templates[view]; ok {
        return
    }

    files := []string{view}
    if len(v.commons) > 0 {
        files = append(files, v.commons...)
    }

    tpl := template.New(filepath.Base(view))

    // add custom func map
    if len(v.funcMap) > 0 {
        tpl.Funcs(v.funcMap)
    }

    // parse template files
    _, e := tpl.ParseFiles(files...)
    if e != nil {
        panic(fmt.Sprintf("failed to parse template, %s, %s", view, e))
    }

    v.templates[view] = tpl
}

func (v *View) normalize(view string) string {
    if ext := filepath.Ext(view); len(ext) == 0 {
        view = view + v.suffix
    }

    if view[:5] != "@view" {
        view = "@view/" + view
    }

    return GetAlias(view)
}
