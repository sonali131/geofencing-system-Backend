# 🚗 Geofencing & Real-Time Vehicle Monitoring System

A production-ready full-stack platform for **real-time vehicle tracking, geofence monitoring, and instant breach detection**. The system leverages geospatial intelligence and WebSocket communication to provide live alerts whenever a vehicle enters or exits predefined zones.

## 🌐 Live Deployment

| Service         | Link                                                |
| --------------- | --------------------------------------------------- |
| 🎨 Frontend     | https://geofencing-system-frontend.vercel.app/      |
| ⚙️ Backend API  | https://geofence-backend-latest.onrender.com        |
| 🐳 Docker Image | https://hub.docker.com/r/sonali131/geofence-backend |

---

## 📸 Project Highlights

✅ Real-time vehicle location tracking

✅ Custom geofence creation and management

✅ Instant breach alerts via WebSockets

✅ High-performance Go backend

✅ MongoDB geospatial queries with 2dsphere indexing

✅ Responsive React dashboard with interactive maps

✅ Automated Docker image deployment through GitHub Actions

---

## 🏗️ System Architecture

Vehicle GPS Data
↓
Go Backend API
↓
MongoDB Atlas
(2dsphere Index)
↓
Geofence Engine
↓
WebSocket Server
↓
React Dashboard

---

## ⚡ Core Features

### 🚨 Real-Time Geofence Alerts

Receive instant notifications whenever a vehicle enters or exits a monitored zone using persistent WebSocket connections.

### 🗺️ Advanced Geospatial Processing

Powered by MongoDB's `$geoIntersects` operator for accurate point-in-polygon calculations and geofence validation.

### 🚙 Vehicle Management

Register, monitor, and manage multiple vehicles through a centralized dashboard.

### 📍 Dynamic Zone Creation

Create and update custom geofence boundaries directly from the interactive map interface.

### 📊 Performance Monitoring

Every API response includes execution metrics (`time_ns`) to help analyze backend performance and latency.

### 🔄 CI/CD Automation

Automated Docker image builds and deployments using GitHub Actions workflows.

---

## 🛠️ Technology Stack

### Backend

- Go (Golang)
- Gorilla WebSockets
- REST APIs

### Database

- MongoDB Atlas
- Geospatial 2dsphere Indexing

### Frontend

- React
- Vite
- Material UI (MUI)
- Leaflet Maps

### DevOps & Infrastructure

- Docker
- GitHub Actions
- Render
- Vercel

---

## 📂 Repository Structure

```text
.
├── backend/
│   ├── go.mod
│   ├── Dockerfile
│   └── main.go
│
├── frontend/
│   ├── src/
│   ├── App.jsx
│   ├── index.css
│   └── package.json
│
└── .github/
    └── workflows/
        └── docker-build.yml
```

---

## 🎯 Use Cases

- Fleet Management Systems
- Logistics & Transportation Tracking
- Asset Monitoring
- Delivery Vehicle Tracking
- Smart Mobility Solutions
- Security & Restricted Area Monitoring

---

## 🔮 Future Enhancements

- SMS & Email Alert Integration
- Historical Route Playback
- Vehicle Analytics Dashboard
- Multi-tenant Architecture
- Role-Based Access Control (RBAC)
- Mobile Application Support

---

### 💡 Why This Project?

This project demonstrates expertise in:

- Real-time Systems
- Geospatial Computing
- WebSocket Communication
- Full-Stack Development
- Cloud Deployment
- Containerization & CI/CD
- Scalable Backend Design

### 👩‍💻 Developer

## Sonali Mishra

Full-Stack Developer passionate about building scalable real-time systems, geospatial applications, and cloud-native solutions.
