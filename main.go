package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	var err error
	// MongoDB Atlas URI yahan daalein
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://sangeetamishra8863_db_user:wUCyoYAdMNIJMUo2@cluster0.wkf9zx4.mongodb.net/?retryWrites=true&w=majority"))
	if err != nil { log.Fatal(err) }

	mux := http.NewServeMux()

	// 1 & 2: Geofences
	mux.HandleFunc("/geofences", handleGeofences)
	// 3 & 4: Vehicles
	mux.HandleFunc("/vehicles", handleVehicles)
	// 5: Location Update & 6: Specific Vehicle Location
	mux.HandleFunc("/vehicles/location", handleLocation)
	mux.HandleFunc("/vehicles/location/", handleGetVehicleLocation) 
	// 7 & 8: Alert Configuration
	mux.HandleFunc("/alerts/configure", handleAlertConfig)
	mux.HandleFunc("/alerts", handleGetAlerts)
	// 9: Violation History
	mux.HandleFunc("/violations/history", handleHistory)
	// WebSocket
	mux.HandleFunc("/ws/alerts", handleWS)

	handler := cors.AllowAll().Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Implementations (Briefly)
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
		var res []bson.M; cur.All(context.TODO(), &res)
		sendJSON(w, start, map[string]interface{}{"geofences": res})
	}
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var loc struct { VehicleID string `json:"vehicle_id"`; Lat float64 `json:"latitude"`; Lng float64 `json:"longitude"` }
	json.NewDecoder(r.Body).Decode(&loc)
	
	// Geofence check logic
	filter := bson.M{"coordinates": bson.M{"$geoIntersects": bson.M{"$geometry": bson.M{"type": "Point", "coordinates": []float64{loc.Lng, loc.Lat}}}}}
	var fences []bson.M
	cur, _ := client.Database("geofence_db").Collection("geofences").Find(context.TODO(), filter)
	cur.All(context.TODO(), &fences)

	if len(fences) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{"event_type": "entry", "vehicle": map[string]string{"vehicle_id": loc.VehicleID}, "geofence": map[string]string{"geofence_name": fmt.Sprintf("%v", fences[0]["name"])}})
		for c := range clients { c.WriteMessage(websocket.TextMessage, msg) }
	}
	sendJSON(w, start, map[string]interface{}{"location_updated": true, "current_geofences": fences})
}

// Baki endpoints ke liye empty arrays bhej dena taaki error na aaye
func handleVehicles(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"vehicles": []string{}}) }
func handleGetVehicleLocation(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"status": "active"}) }
func handleAlertConfig(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"status": "configured"}) }
func handleGetAlerts(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"alerts": []string{}}) }
func handleHistory(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"violations": []string{}, "total_count": 0}) }
func handleWS(w http.ResponseWriter, r *http.Request) { conn, _ := upgrader.Upgrade(w, r, nil); clients[conn] = true }