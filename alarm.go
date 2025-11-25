package main

import (
  "encoding/json"
  "fmt"
  "io"
  "math"
  "net/http"
  "os"
  "strings"
  "sync"
  "time"
)

var (
  latitude        string
  longitude       string
  distance        string
  webhookURL      string
  apiKey          string
  pollInterval    time.Duration
  blacklistTTL    time.Duration
  adsbURL         string
  alertCategories = map[string]bool{}
  blacklist       = make(map[string]int64)
  blacklistMux    sync.Mutex
  quietStart      time.Time
  quietEnd        time.Time
  lastDistance    = make(map[string]float64)
)

type ADSBResponse struct {
  AC []Aircraft `json:"ac"`
}

type Aircraft struct {
  Hex      string  `json:"hex"`
  Category string  `json:"category"`
  Flight   string  `json:"flight"`
  Track    float64 `json:"track"`
  Dir      float64 `json:"dir"`
  Dst      float64 `json:"dst"`
}

// angular diff ≤ 90° → headed toward observer
func isTowardYou(track, dir float64) bool {
  diff := math.Abs(track - dir)
  if diff > 180 {
    diff = 360 - diff
  }
  return diff <= 90
}

// optional: ensure distance decreasing → reduces false positives
func isApproaching(hex string, newDst float64) bool {
  old, exists := lastDistance[hex]
  lastDistance[hex] = newDst
  if !exists {
    return true // first observation
  }
  return newDst < old
}

func isBlacklisted(hex string) bool {
  blacklistMux.Lock()
  defer blacklistMux.Unlock()

  exp, exists := blacklist[hex]
  if !exists {
    return false
  }
  if time.Now().Unix() > exp {
    delete(blacklist, hex)
    return false
  }
  return true
}

func addToBlacklist(hex string) {
  blacklistMux.Lock()
  blacklist[hex] = time.Now().Add(blacklistTTL).Unix()
  blacklistMux.Unlock()
}

func inQuietHours() bool {
  now := time.Now().UTC()

  start := time.Date(
    now.Year(), now.Month(), now.Day(),
    quietStart.Hour(), quietStart.Minute(), 0, 0, time.UTC,
  )

  end := time.Date(
    now.Year(), now.Month(), now.Day(),
    quietEnd.Hour(), quietEnd.Minute(), 0, 0, time.UTC,
  )

  // Handle ranges crossing midnight, e.g., 22:00–06:00
  if end.Before(start) {
    return now.After(start) || now.Before(end)
  }

  return now.After(start) && now.Before(end)
}

func triggerWebhook() {
  client := &http.Client{Timeout: 10 * time.Second}
  req, _ := http.NewRequest("POST", webhookURL, nil)
  req.Header.Set("X-API-KEY", apiKey)

  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("[ERROR] webhook request failed:", err)
    return
  }
  defer func() {
    _ = resp.Body.Close()
  }()

  fmt.Println("[WEBHOOK] Status:", resp.StatusCode)
}

func checkAircraft() {
  client := &http.Client{Timeout: 10 * time.Second}

  req, _ := http.NewRequest("GET", adsbURL, nil)
  req.Header.Set("Accept", "application/json")
  req.Header.Set("X-API-KEY", apiKey)

  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("[ERROR] Failed ADSB request:", err)
    return
  }
  defer func() {
    _ = resp.Body.Close()
  }()

  if resp.StatusCode != 200 {
    fmt.Printf("[ERROR] ADSB returned HTTP %d\n", resp.StatusCode)
    return
  }

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("[ERROR] Failed reading ADSB response:", err)
    return
  }

  if len(body) == 0 {
    fmt.Println("[WARN] ADSB API returned empty body")
    return
  }

  var parsed ADSBResponse
  if err := json.Unmarshal(body, &parsed); err != nil {
    fmt.Println("[ERROR] JSON decode failed:", err)
    fmt.Println("[DEBUG]", string(body))
    return
  }

  var found []string
  for _, a := range parsed.AC {
    cat := strings.ToUpper(strings.TrimSpace(a.Category))
    if !alertCategories[cat] || isBlacklisted(a.Hex) {
      continue
    }
    print("NOT GETTING HERE")

    if !isTowardYou(a.Track, a.Dir) {
      fmt.Printf("[SKIP] %s %s heading away (track %.1f°, dir %.1f°)\n", cat, a.Hex, a.Track, a.Dir)
      continue
    }

    if !isApproaching(a.Hex, a.Dst) {
      fmt.Printf("[SKIP] %s not closing distance (dst %.1f)\n", a.Hex, a.Dst)
      continue
    }

    fmt.Printf("[MATCH] %s %s (%s) inbound — %.1f nm away\n", cat, a.Flight, a.Hex, a.Dst)
    found = append(found, a.Hex)
  }

  if len(found) == 0 {
    return
  }

  if inQuietHours() {
    fmt.Println("[INFO] Quiet hours active — suppressing alert")
  } else {
    fmt.Printf("[ACTION] Triggering webhook for %d inbound aircraft\n", len(found))
    triggerWebhook()
  }

  for _, hex := range found {
    addToBlacklist(hex)
    fmt.Println("[INFO] Added to blacklist:", hex)
  }
}

func main() {
  apiKey = os.Getenv("API_KEY")
  if apiKey == "" {
    fmt.Println("[FATAL] Missing API_KEY env var")
    os.Exit(1)
  }

  latitude = os.Getenv("LATITUDE")
  longitude = os.Getenv("LONGITUDE")
  distance = os.Getenv("DISTANCE")

  if latitude == "" || longitude == "" || distance == "" {
    fmt.Println("[FATAL] LATITUDE, LONGITUDE, and DISTANCE env vars required")
    os.Exit(1)
  }

  webhookURL = os.Getenv("WEBHOOK_URL")
  if webhookURL == "" {
    fmt.Println("[FATAL] WEBHOOK_URL env var required")
    os.Exit(1)
  }

  if v := os.Getenv("QUIET_START"); v != "" {
    if t, err := time.Parse("15:04", v); err == nil {
      quietStart = t
    }
  }

  if v := os.Getenv("QUIET_END"); v != "" {
    if t, err := time.Parse("15:04", v); err == nil {
      quietEnd = t
    }
  }

  if v := os.Getenv("CATEGORIES"); v != "" {
    for _, c := range strings.Split(v, ",") {
      alertCategories[strings.ToUpper(strings.TrimSpace(c))] = true
    }
  }

  if v := os.Getenv("POLL_INTERVAL"); v != "" {
    if d, err := time.ParseDuration(v); err == nil {
      pollInterval = d
    }
  }
  if pollInterval == 0 {
    pollInterval = 30 * time.Second
  }

  if v := os.Getenv("BLACKLIST_TTL"); v != "" {
    if d, err := time.ParseDuration(v); err == nil {
      blacklistTTL = d
    }
  }
  if blacklistTTL == 0 {
    blacklistTTL = 30 * time.Minute
  }

  adsbURL = fmt.Sprintf(
    "https://opendata.adsb.fi/api/v3/lat/%s/lon/%s/dist/%s",
    latitude, longitude, distance,
  )

  fmt.Println("[INFO] ADS-B URL:", adsbURL)
  fmt.Println("[START] ADS-B monitor running...")

  for {
    checkAircraft()
    time.Sleep(pollInterval)
  }
}
