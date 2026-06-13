package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var clients = make(map[*websocket.Conn]bool)

func sendJSON(w http.ResponseWriter, start time.Time, data map[string]interface{}) {
	data["time_ns"] = fmt.Sprintf("%d", time.Since(start).Nanoseconds())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	uri := "mongodb+srv://sangeetamishra8863_db_user:wUCyoYAdMNIJMUo2@cluster0.wkf9zx4.mongodb.net/?retryWrites=true&w=majority"
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil { log.Fatal(err) }

	mux := http.NewServeMux()

	mux.HandleFunc("/geofences", handleGeofences)
	mux.HandleFunc("/vehicles", handleVehicles)
	mux.HandleFunc("/vehicles/location", handleLocation)
	mux.HandleFunc("/vehicles/location/", handleGetVehicleLocation) // Endpoint 6
	mux.HandleFunc("/alerts/configure", handleAlertConfig)          // Endpoint 7
	mux.HandleFunc("/alerts", handleGetAlerts)                      // Endpoint 8
	mux.HandleFunc("/violations/history", handleHistory)            // Endpoint 9
	mux.HandleFunc("/ws/alerts", handleWS)

	handler := cors.AllowAll().Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// 6. GET /vehicles/location/{id}
func handleGetVehicleLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id := strings.TrimPrefix(r.URL.Path, "/vehicles/location/")
	coll := client.Database("geofence_db").Collection("violations")
	var lastV bson.M
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err := coll.FindOne(context.TODO(), bson.M{"vehicle.vehicle_id": id}, opts).Decode(&lastV)
	
	if err != nil {
		sendJSON(w, start, map[string]interface{}{"vehicle_id": id, "status": "unknown"})
		return
	}
	sendJSON(w, start, map[string]interface{}{"vehicle_id": id, "current_location": lastV["location"], "current_geofences": []interface{}{lastV["geofence"]}})
}

// 7. POST /alerts/configure
func handleAlertConfig(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var config bson.M
	json.NewDecoder(r.Body).Decode(&config)
	client.Database("geofence_db").Collection("alert_rules").InsertOne(context.TODO(), config)
	sendJSON(w, start, map[string]interface{}{"status": "configured", "rule": config})
}

// 8. GET /alerts
func handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	cur, _ := client.Database("geofence_db").Collection("alert_rules").Find(context.TODO(), bson.M{})
	var rules []bson.M; cur.All(context.TODO(), &rules)
	sendJSON(w, start, map[string]interface{}{"alerts": rules})
}

// --- Rest of the logic (handleGeofences, handleLocation, handleHistory etc. same as before) ---
func handleGeofences(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	coll := client.Database("geofence_db").Collection("geofences")
	if r.Method == "POST" {
		var in struct { Name string `json:"name"`; Coords [][]float64 `json:"coordinates"`; Cat string `json:"category"` }
		json.NewDecoder(r.Body).Decode(&in)
		var poly [][]float64
		for _, c := range in.Coords { poly = append(poly, []float64{c[1], c[0]}) }
		doc := bson.M{"name": in.Name, "category": in.Cat, "coordinates": bson.M{"type": "Polygon", "coordinates": [][][]float64{poly}}}
		res, _ := coll.InsertOne(context.TODO(), doc)
		sendJSON(w, start, map[string]interface{}{"id": res.InsertedID, "status": "active"})
	} else {
		cur, _ := coll.Find(context.TODO(), bson.M{})
		res := []bson.M{}; cur.All(context.TODO(), &res); sendJSON(w, start, map[string]interface{}{"geofences": res})
	}
}

func handleVehicles(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	coll := client.Database("geofence_db").Collection("vehicles")
	if r.Method == "POST" {
		var v bson.M; json.NewDecoder(r.Body).Decode(&v); res, _ := coll.InsertOne(context.TODO(), v)
		sendJSON(w, start, map[string]interface{}{"id": res.InsertedID, "status": "active"})
	} else {
		cur, _ := coll.Find(context.TODO(), bson.M{}); res := []bson.M{}; cur.All(context.TODO(), &res); sendJSON(w, start, map[string]interface{}{"vehicles": res})
	}
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var loc struct { VehicleID string `json:"vehicle_id"`; Lat float64 `json:"latitude"`; Lng float64 `json:"longitude"` }
	json.NewDecoder(r.Body).Decode(&loc)
	filter := bson.M{"coordinates": bson.M{"$geoIntersects": bson.M{"$geometry": bson.M{"type": "Point", "coordinates": []float64{loc.Lng, loc.Lat}}}}}
	coll := client.Database("geofence_db").Collection("geofences"); var fences []bson.M; cur, err := coll.Find(context.TODO(), filter)
	if err == nil { cur.All(context.TODO(), &fences) }
	if len(fences) > 0 {
		msg := map[string]interface{}{"event_type": "entry", "timestamp": time.Now(), "vehicle": map[string]string{"vehicle_id": loc.VehicleID}, "geofence": map[string]string{"geofence_name": fmt.Sprintf("%v", fences[0]["name"])}, "location": map[string]float64{"latitude": loc.Lat, "longitude": loc.Lng}}
		client.Database("geofence_db").Collection("violations").InsertOne(context.TODO(), msg)
		jsonMsg, _ := json.Marshal(msg)
		for c := range clients { c.WriteMessage(websocket.TextMessage, jsonMsg) }
	}
	sendJSON(w, start, map[string]interface{}{"location_updated": true, "current_geofences": fences})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now(); cur, _ := client.Database("geofence_db").Collection("violations").Find(context.TODO(), bson.M{})
	res := []bson.M{}; cur.All(context.TODO(), &res); sendJSON(w, start, map[string]interface{}{"violations": res})
}

func handleWS(w http.ResponseWriter, r *http.Request) { conn, _ := upgrader.Upgrade(w, r, nil); clients[conn] = true }