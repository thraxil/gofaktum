package faktum

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"http"
	"template"
	"strings"
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
	Fact *datastore.Key
	Tag *datastore.Key
}

// routing
func init() {
	http.HandleFunc("/",index)
	http.HandleFunc("/login/",login)
	http.HandleFunc("/add/",add)
}

type PageData struct {
	Title string
	Facts []FactWithKey
	SourceURL string
	SourceName string
	Details string
	FactTitle string
	User string
}

// model + key data structures
type FactWithKey struct {
	Key        string
	Title      string
	Details    string
	SourceUrl  string
	SourceName string
	AddDate    datastore.Time
	User       string
}

// controller functions

var indexTmpl = template.Must(template.New("index").Funcs(
	template.FuncMap{"convertToTime": 
	func(t datastore.Time) *time.Time { 
                return t.Time() 
}}).ParseFile("templates/index.html"))

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	q := datastore.NewQuery("Fact").Order("-AddDate").Limit(10)
	facts := make([]FactWithKey, 0, 10)
	for t := q.Run(c); ; {
		var x Fact
		key, err := t.Next(&x)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
		fk := FactWithKey{
		Key: key.String(),
		Title: x.Title,
		SourceUrl: x.SourceUrl,
		SourceName: x.SourceName,
		Details: x.Details,
		User: x.User,
		}
		facts = append(facts,fk)
	}
	p := PageData{
	Title: "Faktum",
	Facts: facts,
	SourceURL: r.FormValue("source_url"),
	SourceName: r.FormValue("source_name"),
	Details: r.FormValue("details"),
	FactTitle: r.FormValue("title"),
	User: u.String(),
	}

	if err := indexTmpl.Execute(w, p); err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
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
	}
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func addTagToFact(c appengine.Context, factKey *datastore.Key, tag string) {
	q := datastore.NewQuery("Tag").Filter("Name =", tag)
	cnt,_ := q.Count(c)
	var existingTag Tag
	var existingTagKey *datastore.Key
	if cnt > 0 {
		// retrieve
		for t := q.Run(c); ; {
			existingTagKey, _ = t.Next(&existingTag)
			break // only need one
		}
	} else {
		// create a new one
		var t = Tag{Name: tag}
		existingTagKey, _ = datastore.Put(c, datastore.NewIncompleteKey(c, "Tag", nil), &t)
	}
	qft := datastore.NewQuery("FactTag").
		Filter("Fact = ", factKey).
		Filter("Tag = ", existingTagKey)
	cnt2,_ := qft.Count(c)
	if cnt2 == 0 {
		// create a new one
		var ft = FactTag{Fact: factKey, Tag: existingTagKey}
		_, _ = datastore.Put(c, datastore.NewIncompleteKey(c, "FactTag", nil), &ft)
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
        Title:      r.FormValue("title"),
	Details:    r.FormValue("details"),
	SourceUrl:  r.FormValue("source_url"),
	SourceName: r.FormValue("source_name"),
        AddDate:    datastore.SecondsToTime(time.Seconds()),
	User:       u.String(),
	}
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Fact", nil), &f)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	
	tags := strings.Split(r.FormValue("tags"),",")
	for _,t := range(tags) {
		t = strings.Trim(t," ")
		fmt.Printf("<%q>\n", t)
		addTagToFact(c,key,t)
	}
	

	http.Redirect(w, r, "/", http.StatusFound)
}
