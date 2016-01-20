package db

import (
	"database/sql"
	"errors"
	"github.com/go-on/wsi"
	_ "gopkg.in/go-on/pq.v2"
	"net/http"
	"strconv"
)

type db struct {
	db *sql.DB
}

func New(d *sql.DB) *db {
	return &db{d}
}

// github.com/go-on/wsi

var testData = []map[string]wsi.Setter{
	map[string]wsi.Setter{"ID": wsi.SetInt(12), "Name": wsi.SetString("Adrian")},
	map[string]wsi.Setter{"ID": wsi.SetInt(24), "Name": wsi.SetString("George")},
}

func (d *db) SearchPersons(limit, offset int, w http.ResponseWriter, r *http.Request) (wsi.Scanner, error) {
	return wsi.NewTestQuery([]string{"ID", "Name"}, testData...), nil
	/*
		if len(opt.OrderBy) == 0 {
			opt.OrderBy = append(opt.OrderBy, "id asc")
		}
		return wsi.DBQuery(d.db, "select id,name from person order by $1 limit $2, $3", strings.Join(opt.OrderBy, ","), opt.Offset, opt.Limit)
	*/
}

func (d *db) ReadPerson(limit, offset int, w http.ResponseWriter, r *http.Request) (wsi.Scanner, error) {
	id, err := strconv.Atoi(r.URL.Fragment)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID is not a number"))
		return nil, err
	}

	switch id {
	case 12:
		return wsi.NewTestQuery([]string{"ID", "Name"}, testData[0]), nil
	case 24:
		return wsi.NewTestQuery([]string{"ID", "Name"}, testData[1]), nil
	default:
		err = errors.New("not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return nil, err
	}
}

func (d *db) CreatePerson(m map[string]interface{}, w http.ResponseWriter, r *http.Request) error {
	// we fake a created response here
	res := map[string]interface{}{"ID": 400, "Name": m["Name"]}
	w.WriteHeader(http.StatusCreated)
	wsi.ServeJSON(res, w)
	return nil
}

func (d *db) UpdatePerson(m map[string]interface{}, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Fragment)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID is not a number"))
		return err
	}

	switch id {
	case 12:
		m["ID"] = 12
	case 24:
		m["ID"] = 24
	default:
		err = errors.New("not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return err
	}
	// we fake a created response here
	w.WriteHeader(http.StatusOK)
	wsi.ServeJSON(m, w)
	return nil
}

func (d *db) ReplacePerson(m map[string]interface{}, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Fragment)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID is not a number"))
		return err
	}

	switch id {
	case 12:
		m["ID"] = 12
	case 24:
		m["ID"] = 24
	default:
		err = errors.New("not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return err
	}
	// we fake a created response here
	w.WriteHeader(http.StatusOK)
	wsi.ServeJSON(m, w)
	return nil
}

func (d *db) DeletePerson(m map[string]interface{}, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.URL.Fragment)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID is not a number"))
		return err
	}

	switch id {
	case 12:
		m["ID"] = 12
	case 24:
		m["ID"] = 24
	default:
		err = errors.New("not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return err
	}
	w.WriteHeader(http.StatusGone)
	w.Write([]byte(`deleted`))
	return nil
}
