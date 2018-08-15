package main

import (
	"fmt"
	"net/http"

	"github.com/go-on/router/route/new_idea/db"
	"github.com/go-on/router/route/new_idea/person"
	"github.com/go-on/router/route/new_idea/router"
	"github.com/metakeule/dbwrap"
)

/*

das problem / die vereinfachung:

wenn wir auf strenges REST verzichten, können wir folgende vereinfachungen vornehmen:


- url-parameter nur in GET urls, sonst (POST,PUT,PATCH,DELETE) immer ohne (d.h. wert wird im body mitgegeben)
- POST,PUT,PATCH,DELETE: body immer JSON
- jede GET route braucht immer nur einen platzhalter, dieser muss auch uniq sein für die routen,
  daher bietet sich ein automatisch generierter platzhalter an. der name der route wird durch den typ bestimmt.
- die get route kann auf einem typen definiert werden; dieser typ bestimmt automatisch die konvertierung
- fehlerhafte GET parameter können auf der router ebene behandelt werden (entsprechende fehlerstati werden zurückgeliefert),
das gleiche kann für fehlerhafte jsons gemacht werden. hat den vorteil, dass validierung (d.h. userfeedback) von malicious dingen (angriff)
sauber getrennt werden kann. die malicious sachen werden schon vom router abgefangen und bekommen nur einsprechende http status codes,
die fehlerhaften validierungen bekommen zusätzlich einen body mit hinweisen (dieser body ist anwendungspezifisch)
- generelle get suchanfragen mit entsprechenden parametern können auch zentral abgebildet werden (paging, limit, skip, sort)
- es gibt nur get routen ohne parameter und welche mit einem parameter, also such routen und ressourcen routen
- um unterressourcen zu handeln, statt /person/:person_id/address/:address_id entweder
     /address/:address_id (wenn die id universell ist) oder
     /person-address/:person_id-address_id , so dass der parameter wert beides beinhaltet und dann kann man mit split
     oder regulären ausdrücken die zwei oder mehreren elemente rausholen. auf diese weise bleiben die urls tendenziell "flach" und der router
     muss immer alles durchsuchen  (es gibt z.b. keinen konflikt zwischen einer route und einem parameter, wie bei /person/address/:person_id-address_id vs
     /person/:person_id). das heisst im klartext: in einem prefix ist kein / erlaubt und für den router bedeutet es:
     er muss nur einmal nach subroutern prüfen, wenn es kein subrouter ist, kann er im entsprechenden map der entsprechenden methode nachschauen
     und wenn es bei get nicht direkt mapt und einen slash beinhaltet, ist alles hinter dem slash der parameter: ganz einfach

*/

func printErr(r *http.Request, err error) {
	fmt.Printf("error in %s %s: %s\n", r.Method, r.URL.Path, err)
}

func main() {

	_, DB := dbwrap.NewFake()

	d := db.New(DB)

	p := person.New()
	p.Create = d.CreatePerson
	p.Read = d.ReadPerson
	p.Update = d.UpdatePerson
	p.Delete = d.DeletePerson
	p.Index = d.SearchPersons
	p.Replace = d.ReplacePerson
	p.ErrorHandler = printErr

	rt := router.New()

	personsRoute, personRoute := rt.RouteRessource(p, "person")
	m := router.New()
	rt.SubRouter("sub", m)
	/*
		getPerson := rt.RouteFuncParam(method.GET, "person", person.Get)
		indexPerson := rt.RouteFunc(method.GET, "person", person.Index)
		postPerson := rt.RouteFunc(method.POST, "person", person.Index)
	*/

	// fmt.Printf("%#v\n", m)
	// fmt.Printf("%#v\n", rt)
	fmt.Println(`MountPoint():`, rt.MountPoint())
	fmt.Println(`getPerson("24"): `, personRoute("24"))
	fmt.Println("postPerson(): ", personsRoute())
	fmt.Println("searchPerson(): ", personsRoute())

	http.ListenAndServe(":8085", m)
}
