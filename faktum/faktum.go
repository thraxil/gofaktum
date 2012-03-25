package facktum

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"http"
	"template"
	"time"
)

// model data structures
type Fact struct {
	Title      string
	Details    string
	SourceUrl  string
	SourceName string
	AddDate    datastore.Time
	User       string
}

type Tag struct {
	Name string
}

type FactTag struct {
	Fact string
	Tag string
}

// routing
func init() {
	http.HandleFunc("/",index)
	http.HandleFunc("/add/",add)
}

type PageData struct {
	Title string
	Facts []Fact
}

// controller functions

var indexTmpl = template.Must(template.New("index").ParseFile("templates/index.html"))

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Fact").Order("-AddDate").Limit(10)
	facts := make([]Fact, 0, 10)
	if _, err := q.GetAll(c, &facts); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	p := PageData{
	Title: "Faktum",
	Facts: facts,
	}

	if err := indexTmpl.Execute(w, p); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	f := Fact{
        Title: r.FormValue("title"),
	Details: r.FormValue("details"),
	SourceUrl: r.FormValue("source_url"),
	SourceName: r.FormValue("source_name"),
        AddDate:    datastore.SecondsToTime(time.Seconds()),
	}
	if u := user.Current(c); u != nil {
		f.User = u.String()
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Fact", nil), &f)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
