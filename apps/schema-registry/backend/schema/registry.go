package schema

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/eagraf/habitat/structs/sources"
	"github.com/gorilla/mux"
)

// Registry is a server that handles all state for storing schemas on a node in a community
type Registry struct {
	store sources.SchemaStore
}

func NewRegistry(store sources.SchemaStore) *Registry {
	return &Registry{
		store: store,
	}
}

func (r *Registry) AddSchemaHandler(w http.ResponseWriter, req *http.Request) {

	var addReq ctl.AddSchemaRequest
	err := ctl.RequestFromBody(req, &addReq)

	if err != nil {
		http.Error(w, fmt.Sprintf("json unmarshal failed with %s", err.Error()), http.StatusBadRequest)
	}

	err = r.store.Add(addReq.Sch)
	if err != nil {
		http.Error(w, fmt.Sprintf("error adding source schema %s", err.Error()), http.StatusInternalServerError)
	}
	api.WriteResponse(w, ctl.AddSchemaResponse{
		ID: addReq.Sch.ID,
	})

}

func (r *Registry) GetSchemaHandler(w http.ResponseWriter, req *http.Request) {

	var getReq ctl.GetSchemaRequest
	err := ctl.RequestFromBody(req, &getReq)

	if err != nil {
		http.Error(w, fmt.Sprintf("json unmarshal failed with %s", err.Error()), http.StatusBadRequest)
	}

	sch, err := r.store.Get(getReq.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error looking up source schema %s", err.Error()), http.StatusInternalServerError)
	}

	api.WriteResponse(w, ctl.GetSchemaResponse{
		Sch: sch,
	})
}

func (r *Registry) DeleteSchemaHandler(w http.ResponseWriter, req *http.Request) {
	var delReq ctl.DeleteSchemaRequest
	err := ctl.RequestFromBody(req, &delReq)

	if err != nil {
		http.Error(w, fmt.Sprintf("json unmarshal failed with %s", err.Error()), http.StatusBadRequest)
	}

	err = r.store.Delete(delReq.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error looking up source schema %s", err.Error()), http.StatusInternalServerError)
	}

	api.WriteResponse(w, ctl.DeleteSchemaResponse{})
}

func (r *Registry) Serve(port string) error {
	router := mux.NewRouter()

	router.HandleFunc("/add_schema", r.AddSchemaHandler)
	router.HandleFunc("/get_schema", r.GetSchemaHandler)
	router.HandleFunc("/delete_schema", r.DeleteSchemaHandler)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return srv.ListenAndServe()
}
