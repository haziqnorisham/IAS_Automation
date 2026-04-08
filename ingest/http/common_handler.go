package http

import (
	"fmt"
	ias_pg "ias/automation/db/pg"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ias_pg.NewPostgresStorage(nil).QueryData("select device_name from ppj_tree_sensor")
	fmt.Fprint(w, "Hello World!")
}
