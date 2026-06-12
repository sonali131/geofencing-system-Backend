// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/gorilla/websocket"
// 	"github.com/rs/cors"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// var client *mongo.Client
// var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
// var clients = make(map[*websocket.Conn]bool)

// func sendJSON(w http.ResponseWriter, start time.Time, data map[string]interface{}) {
// 	data["time_ns"] = fmt.Sprintf("%d", time.Since(start).Nanoseconds())
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(data)
// }

// func main() {
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	var err error
// 	// MongoDB Atlas URI
// 	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://sangeetamishra8863_db_user:wUCyoYAdMNIJMUo2@cluster0.wkf9zx4.mongodb.net/?retryWrites=true&w=majority"))
// 	if err != nil { log.Fatal(err) }

// 	mux := http.NewServeMux()

// 	mux.HandleFunc("/geofences", handleGeofences)
// 	mux.HandleFunc("/vehicles", handleVehicles)
// 	mux.HandleFunc("/vehicles/location", handleLocation)
// 	mux.HandleFunc("/ws/alerts", handleWS)
// 	// Baki handlers ko stub kar diya taaki crash na ho
// 	mux.HandleFunc("/violations/history", handleHistory)
// 	mux.HandleFunc("/alerts", handleGetAlerts)

// 	fmt.Println("Backend started on :8080")
// 	handler := cors.AllowAll().Handler(mux)
// 	log.Fatal(http.ListenAndServe(":8080", handler))
// }

// func handleGeofences(w http.ResponseWriter, r *http.Request) {
// 	start := time.Now()
// 	coll := client.Database("geofence_db").Collection("geofences")
// 	if r.Method == "POST" {
// 		var in struct { Name string `json:"name"`; Coords [][]float64 `json:"coordinates"`; Cat string `json:"category"` }
// 		json.NewDecoder(r.Body).Decode(&in)
		
// 		var poly [][]float64
// 		// Frontend [Lat, Lng] -> MongoDB [Lng, Lat]
// 		for _, c := range in.Coords { poly = append(poly, []float64{c[1], c[0]}) }
		
// 		doc := bson.M{
// 			"name": in.Name, 
// 			"category": in.Cat, 
// 			"coordinates": bson.M{"type": "Polygon", "coordinates": [][][]float64{poly}},
// 		}
// 		res, _ := coll.InsertOne(context.TODO(), doc)
// 		sendJSON(w, start, map[string]interface{}{"id": res.InsertedID, "status": "active"})
// 	} else {
// 		cur, _ := coll.Find(context.TODO(), bson.M{})
// 		res := []bson.M{} // Initialize as empty slice
// 		cur.All(context.TODO(), &res)
// 		sendJSON(w, start, map[string]interface{}{"geofences": res})
// 	}
// }

// func handleLocation(w http.ResponseWriter, r *http.Request) {
// 	start := time.Now()
// 	var loc struct { VehicleID string `json:"vehicle_id"`; Lat float64 `json:"latitude"`; Lng float64 `json:"longitude"` }
// 	json.NewDecoder(r.Body).Decode(&loc)
	
// 	// FIX: Fences ko hamesha empty array rakhein, null nahi
// 	fences := []bson.M{}
	
// 	// MongoDB GeoQuery: Point expects [Longitude, Latitude]
// 	filter := bson.M{
// 		"coordinates": bson.M{
// 			"$geoIntersects": bson.M{
// 				"$geometry": bson.M{
// 					"type":        "Point",
// 					"coordinates": []float64{loc.Lng, loc.Lat}, 
// 				},
// 			},
// 		},
// 	}

// 	coll := client.Database("geofence_db").Collection("geofences")
// 	cur, err := coll.Find(context.TODO(), filter)
// 	if err == nil {
// 		cur.All(context.TODO(), &fences)
// 	}

// 	// Alert logic
// 	if len(fences) > 0 {
// 		gName := "Safe Zone"
// 		if name, ok := fences[0]["name"].(string); ok { gName = name }
		
// 		msg, _ := json.Marshal(map[string]interface{}{
// 			"event_type": "entry",
// 			"vehicle":    map[string]string{"vehicle_id": loc.VehicleID},
// 			"geofence":   map[string]string{"geofence_name": gName},
// 		})
// 		for c := range clients {
// 			c.WriteMessage(websocket.TextMessage, msg)
// 		}
// 	}

// 	sendJSON(w, start, map[string]interface{}{
// 		"location_updated":  true, 
// 		"current_geofences": fences,
// 	})
// }

// func handleWS(w http.ResponseWriter, r *http.Request) {
// 	conn, _ := upgrader.Upgrade(w, r, nil)
// 	clients[conn] = true
// }

// // Stubs
// func handleVehicles(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"vehicles": []bson.M{}}) }
// func handleGetVehicleLocation(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"status": "active"}) }
// func handleHistory(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"violations": []string{}}) }
// func handleGetAlerts(w http.ResponseWriter, r *http.Request) { sendJSON(w, time.Now(), map[string]interface{}{"alerts": []string{}}) }
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var clients = make(map[*websocket.Conn]bool)

// Helper: Strict JSON Response with time_ns
func sendJSON(w http.ResponseWriter, start time.Time, data map[string]interface{}) {
	data["time_ns"] = fmt.Sprintf("%d", time.Since(start).Nanoseconds())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var err error
	uri := "mongodb+srv://sangeetamishra8863_db_user:wUCyoYAdMNIJMUo2@cluster0.wkf9zx4.mongodb.net/?retryWrites=true&w=majority"
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil { log.Fatal(err) }

	mux := http.NewServeMux()

	mux.HandleFunc("/geofences", handleGeofences)
	mux.HandleFunc("/vehicles", handleVehicles) // Updated
	mux.HandleFunc("/vehicles/location", handleLocation) // Updated
	mux.HandleFunc("/violations/history", handleHistory) // Updated
	mux.HandleFunc("/ws/alerts", handleWS)

	fmt.Println("Backend Live on :8080")
	handler := cors.AllowAll().Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// 1. GEOFENCES (POST/GET)
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
		res := []bson.M{}; cur.All(context.TODO(), &res)
		sendJSON(w, start, map[string]interface{}{"geofences": res})
	}
}

// 2. VEHICLES (POST/GET)
func handleVehicles(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	coll := client.Database("geofence_db").Collection("vehicles")
	if r.Method == "POST" {
		var v bson.M
		json.NewDecoder(r.Body).Decode(&v)
		v["created_at"] = time.Now()
		res, _ := coll.InsertOne(context.TODO(), v)
		sendJSON(w, start, map[string]interface{}{"id": res.InsertedID, "status": "active"})
	} else {
		cur, _ := coll.Find(context.TODO(), bson.M{})
		res := []bson.M{}; cur.All(context.TODO(), &res)
		sendJSON(w, start, map[string]interface{}{"vehicles": res})
	}
}

// 3. LOCATION UPDATE & VIOLATION SAVING
func handleLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var loc struct { VehicleID string `json:"vehicle_id"`; Lat float64 `json:"latitude"`; Lng float64 `json:"longitude"` }
	json.NewDecoder(r.Body).Decode(&loc)
	
	fences := []bson.M{}
	filter := bson.M{"coordinates": bson.M{"$geoIntersects": bson.M{"$geometry": bson.M{"type": "Point", "coordinates": []float64{loc.Lng, loc.Lat}}}}}
	coll := client.Database("geofence_db").Collection("geofences")
	cur, err := coll.Find(context.TODO(), filter)
	if err == nil { cur.All(context.TODO(), &fences) }

	if len(fences) > 0 {
		gName := fmt.Sprintf("%v", fences[0]["name"])
		msg := map[string]interface{}{
			"event_type": "entry",
			"timestamp":  time.Now(),
			"vehicle":    map[string]string{"vehicle_id": loc.VehicleID},
			"geofence":   map[string]string{"geofence_name": gName},
			"location":   map[string]float64{"latitude": loc.Lat, "longitude": loc.Lng},
		}

		// ZAROORI: Violation ko Database mein save karein
		client.Database("geofence_db").Collection("violations").InsertOne(context.TODO(), msg)

		// WebSocket Broadcast
		jsonMsg, _ := json.Marshal(msg)
		for c := range clients { c.WriteMessage(websocket.TextMessage, jsonMsg) }
	}

	sendJSON(w, start, map[string]interface{}{"location_updated": true, "current_geofences": fences})
}

// 4. VIOLATION HISTORY (GET)
func handleHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	coll := client.Database("geofence_db").Collection("violations")
	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(50) // Latest 50 violations
	cur, _ := coll.Find(context.TODO(), bson.M{}, opts)
	res := []bson.M{}; cur.All(context.TODO(), &res)
	sendJSON(w, start, map[string]interface{}{"violations": res, "total_count": len(res)})
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	clients[conn] = true
}

func handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, time.Now(), map[string]interface{}{"alerts": []string{}})
}