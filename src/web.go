package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//go:embed templates/index.html
var indexHTML string

//go:embed static/*
var StaticFiles embed.FS

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Log HTTP request
	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)

	// Retrieve cookies
	cityNameCookie, _ := r.Cookie("cityName")
	latCookie, _ := r.Cookie("latitude")
	lonCookie, _ := r.Cookie("longitude")

	cityName := ""
	latitude := ""
	longitude := ""

	// Set defaults if cookies exist
	if cityNameCookie != nil {
		cityName = cityNameCookie.Value
	}
	if latCookie != nil {
		latitude = latCookie.Value
	}
	if lonCookie != nil {
		longitude = lonCookie.Value
	}

	// Insert cookies into JavaScript for the frontend
	pageWithCookies := fmt.Sprintf(indexHTML, cityName, latitude, longitude)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, pageWithCookies)
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	if lat == "" || lon == "" {
		http.Error(w, "Latitude and longitude are required", http.StatusBadRequest)
		return
	}

	latitude, err1 := strconv.ParseFloat(lat, 64)
	longitude, err2 := strconv.ParseFloat(lon, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid latitude or longitude", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Requested weather data for lat: %s, lon: %s", lat, lon)

	data := OpenMeteoAPIResponse{}
	data.FetchData(OpenMeteoAPIEndpoint, OpenMeteoAPIParams, float64ToSting(latitude), float64ToSting(longitude))
	weatherTable := data.Points().setMoonIllumination().setSeeing().Print()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, weatherTable)
}

func handleSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	query = strings.TrimSpace(query)

	if query == "" {
		http.Error(w, "Query parameter is missing", http.StatusBadRequest)
		return
	}

	suggestions, err := fetchSuggestions(query)
	if err != nil {
		http.Error(w, "Unable to fetch suggestions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

func handleRobots(w http.ResponseWriter, r *http.Request) {
	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)

	robotsTxtFile, _ := StaticFiles.ReadFile("static/robots.txt")
	contentType := "text/plain"

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(robotsTxtFile)
	if err != nil {
		log.Printf("ERROR: Could not write robots.txt: %v", err)
	}
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	log.Printf("INFO: From: %s | User-Agent: %s | Path: %s", r.RemoteAddr, r.UserAgent(), r.URL.Path)

	faviconFile, _ := StaticFiles.ReadFile("static/favicon.ico")
	contentType := "image/x-icon"

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(faviconFile)
	if err != nil {
		log.Printf("ERROR: Could not write favicon.ico: %v", err)
	}
}

func float64ToSting(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}
