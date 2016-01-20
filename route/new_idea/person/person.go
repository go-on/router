package person

import (
	"errors"
	"fmt"
	"github.com/go-on/wsi"
	"net/http"
)

type Ressource struct {
	Create  wsi.ExecFunc
	Read    wsi.QueryFunc
	Update  wsi.ExecFunc
	Delete  wsi.ExecFunc
	Replace wsi.ExecFunc
	Index   wsi.QueryFunc
	wsi.Ressource
}

type Person struct {
	ID        int
	Name      string
	Ressource `json:"-"`
}

func (p Person) POST(w http.ResponseWriter, r *http.Request)   { p.ServeExec(p.Create, w, r) }
func (p Person) GET(w http.ResponseWriter, r *http.Request)    { p.ServeQuery(p.Read, w, r) }
func (p Person) PATCH(w http.ResponseWriter, r *http.Request)  { p.ServeExec(p.Update, w, r) }
func (p Person) DELETE(w http.ResponseWriter, r *http.Request) { p.ServeExec(p.Delete, w, r) }
func (p Person) PUT(w http.ResponseWriter, r *http.Request)    { p.ServeExec(p.Replace, w, r) }
func (p Person) INDEX(w http.ResponseWriter, r *http.Request)  { p.ServeQuery(p.Index, w, r) }

func (p Person) ValidatePUT() map[string]error   { return p.validate() }
func (p Person) ValidatePOST() map[string]error  { return p.validate() }
func (p Person) ValidatePATCH() map[string]error { return p.validate() }

func (p Person) validate() map[string]error {
	errs := map[string]error{}
	if p.Name == "" {
		errs["Name"] = errors.New("name must not be empty")
	}
	return errs
}

func mkPerson() interface{} { return &Person{} }

func printErr(r *http.Request, err error) {
	fmt.Printf("Error in route GET %s: %T %s\n", r.URL.Path, err, err.Error())
}

func New() (p Person) {
	p.Ressource.Ressource.RessourceFunc = mkPerson
	return
}

// ensure the interfaces are fulfilled
// var _ wsi.ColumnsMapper = &Person{}
var _ wsi.PUTValidater = &Person{}
var _ wsi.POSTValidater = &Person{}
var _ wsi.PATCHValidater = &Person{}
