package main

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// database file
const Database = "./database.json"

type readOp struct {
	key    string
	filter string
	resp   chan map[string]string
}

type writeOp struct {
	datas map[string]string
	resp  chan bool
}

var reads = make(chan *readOp)
var writes = make(chan *writeOp)

// Decode json string
func JsonDecode(s []byte) map[string]string {
	var parsed map[string]string
	err := json.Unmarshal(s, &parsed)
	if err != nil {
		log.Println(err)
		return nil
	}
	return parsed
}

// Encode map to json string
func JsonEncode(datas map[string]string) []byte {
	str, err := json.Marshal(datas)
	if err != nil {
		log.Println(err)
		return nil
	}
	return str
}

// Post datas Handler
func SetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// POST takes the data and saves it to disk.
	case "POST":
		// get data input field
		str := r.FormValue("data")
		if str == "" {
			http.Error(w, "no data input field.", http.StatusInternalServerError)
			return
		}

		// decode datas
		keys := JsonDecode([]byte(str))
		if keys == nil {
			http.Error(w, "could not parse json data.", http.StatusInternalServerError)
			return
		}

		// update data to db
		write := &writeOp{
			datas: keys,
			resp:  make(chan bool)}

		writes <- write
		<-write.resp

		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET JSON output
func GetJsonHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// GET json datas
	case "GET":
		//filter := r.URL.Query().Get("filter")
		//key := r.URL.Query().Get("key")

		read := &readOp{
			key:    "",
			filter: "",
			resp:   make(chan map[string]string)}

		reads <- read
		datas := <-read.resp

		w.Header().Set("Content-Type", "application/json")
		str := JsonEncode(datas)
		w.Write(str)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET HTML output
func GetHtmlHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// GET json datas
	case "GET":

		read := &readOp{
			key:    "",
			filter: "",
			resp:   make(chan map[string]string)}

		reads <- read
		datas := <-read.resp

		w.Header().Set("Content-Type", "text/html")

		htmltemplate := `<html><body><ul>
      {{range $key, $value := .}}<li>{{$key}} : {{$value}}</li>{{end}}
      </body></ul></html>`

		t := template.New("t")
		t, err := t.Parse(htmltemplate)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, datas)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// GET CSV output
func GetCsvHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// GET csv datas
	case "GET":
		read := &readOp{
			key:    "",
			filter: "",
			resp:   make(chan map[string]string)}

		reads <- read
		datas := <-read.resp

		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		for key, value := range datas {
			record := []string{key, value}
			err := writer.Write(record)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		writer.Flush()
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Read Json Database from filesystem
func DatbaseLoad(db string) map[string]string {
	content, err := ioutil.ReadFile(db)
	if err != nil {
		return make(map[string]string)
	}
	return JsonDecode(content)
}

// Dump data to File
func DataBaseDump(db string, datas map[string]string) {
	file, err := os.Create(db)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	str := JsonEncode(datas)
	if str != nil {
		file.Write([]byte(str))
	}
	file.Close()
}

// Go routine handling ios
func DBHandler(db string) {
	// load database
	Datas := DatbaseLoad(Database)
	for {
		// handle io from handlers
		select {
		// get
		case read := <-reads:
			if read.key != "" {
				var rc = make(map[string]string)
				rc[read.key] = Datas[read.key]
				read.resp <- rc
			} else if read.filter != "" {
				//
			} else {
				read.resp <- Datas
			}
		// set
		case write := <-writes:
			for key, value := range write.datas {
				Datas[key] = value
			}
			log.Printf("dump database to %s\n", db)
			DataBaseDump(db, Datas)
			write.resp <- true
		}
	}
}

func main() {
	// routine to handle writes to filesystem
	go DBHandler(Database)
	// http routes
	http.HandleFunc("/set", SetHandler)
	http.HandleFunc("/get", GetJsonHandler)
	http.HandleFunc("/get/", GetJsonHandler)
	http.HandleFunc("/get/html", GetHtmlHandler)
	http.HandleFunc("/get/csv", GetCsvHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
