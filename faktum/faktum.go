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
	SourceURL string
	SourceName string
	Details string
	FactTitle string
}

// controller functions

var indexTmpl = template.Must(template.New("index").Funcs(
	template.FuncMap{"convertToTime": 
	func(t datastore.Time) *time.Time { 
                return t.Time() 
}}).ParseFile("templates/index.html"))

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
	SourceURL: r.FormValue("source_url"),
	SourceName: r.FormValue("source_name"),
	Details: r.FormValue("details"),
	FactTitle: r.FormValue("title"),
	}

	if err := indexTmpl.Execute(w, p); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}

	f := Fact{
        Title: r.FormValue("title"),
	Details: r.FormValue("details"),
	SourceUrl: r.FormValue("source_url"),
	SourceName: r.FormValue("source_name"),
        AddDate:    datastore.SecondsToTime(time.Seconds()),
	User: u.String(),
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Fact", nil), &f)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
