package http

import (
	"fmt"
	ias_pg "ias/automation/db/pg"
	"net/http"
)

func getAllTreeSensorHandler(w http.ResponseWriter, r *http.Request) {

	sensors, _ := ias_pg.NewPostgresStorage(nil).QueryData("select * from ppj_tree_sensor")
	for _, sensor := range sensors {
		fmt.Fprintf(w, "Sensor: %+v\n", sensor)
	}
}
