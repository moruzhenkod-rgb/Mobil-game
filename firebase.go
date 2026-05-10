package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ── Config (replace with real project values before shipping) ─────

const (
	fbAPIKey     = "AIzaSyPLACEHOLDER_REPLACE_ME"
	fbProjectID  = "runic-crush-game"
	fbAuthBase   = "https://identitytoolkit.googleapis.com/v1"
	fbStoreBase  = "https://firestore.googleapis.com/v1/projects/" + fbProjectID + "/databases/(default)/documents"
)

// ── Client ────────────────────────────────────────────────────────

type FirebaseClient struct {
	http    *http.Client
	idToken string
	uid     string
	ready   bool
	offline bool
}

var fb = &FirebaseClient{
	http: &http.Client{Timeout: 8 * time.Second},
}

// InitFirebase should be called once at startup (in a goroutine).
func InitFirebase() {
	if fbAPIKey == "AIzaSyPLACEHOLDER_REPLACE_ME" {
		fb.offline = true
		return
	}
	go func() {
		if err := fb.signInAnonymously(); err != nil {
			fb.offline = true
			return
		}
		fb.ready = true
		fb.syncSave()
	}()
}

func (f *FirebaseClient) IsReady() bool  { return f.ready }
func (f *FirebaseClient) IsOffline() bool { return f.offline }

// ── Auth ──────────────────────────────────────────────────────────

type authResponse struct {
	IDToken      string `json:"idToken"`
	LocalID      string `json:"localId"`
	RefreshToken string `json:"refreshToken"`
}

func (f *FirebaseClient) signInAnonymously() error {
	body, _ := json.Marshal(map[string]bool{"returnSecureToken": true})
	url := fmt.Sprintf("%s/accounts:signInAnonymously?key=%s", fbAuthBase, fbAPIKey)
	resp, err := f.post(url, body)
	if err != nil {
		return err
	}
	var auth authResponse
	if err := json.Unmarshal(resp, &auth); err != nil {
		return err
	}
	f.idToken = auth.IDToken
	f.uid = auth.LocalID
	return nil
}

// SignInWithGoogle uses a Google ID token obtained from platform SDK.
func (f *FirebaseClient) SignInWithGoogle(googleToken string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"postBody":          "id_token=" + googleToken + "&providerId=google.com",
		"requestUri":        "http://localhost",
		"returnSecureToken": true,
	})
	url := fmt.Sprintf("%s/accounts:signInWithIdp?key=%s", fbAuthBase, fbAPIKey)
	resp, err := f.post(url, body)
	if err != nil {
		return err
	}
	var auth authResponse
	if err := json.Unmarshal(resp, &auth); err != nil {
		return err
	}
	f.idToken = auth.IDToken
	f.uid = auth.LocalID
	f.ready = true
	return nil
}

// ── Cloud Save (Firestore) ────────────────────────────────────────

type firestoreDoc struct {
	Fields map[string]interface{} `json:"fields"`
}

func (f *FirebaseClient) syncSave() {
	if f.uid == "" {
		return
	}
	remote, err := f.loadRemoteSave()
	if err != nil {
		return // silent fail, offline
	}
	// Merge: higher UnlockedLevel wins
	if remote.UnlockedLevel > progress.UnlockedLevel {
		progress.UnlockedLevel = remote.UnlockedLevel
	}
	for i, s := range remote.Stars {
		if s > progress.Stars[i] {
			progress.Stars[i] = s
		}
	}
	progress.Coins += remote.Coins // merge coins (both sides earned independently)
	saveProgress()
}

type remoteData struct {
	UnlockedLevel int
	Stars         [MaxLevels + 1]int
	Coins         int
}

func (f *FirebaseClient) loadRemoteSave() (remoteData, error) {
	url := fmt.Sprintf("%s/users/%s", fbStoreBase, f.uid)
	body, err := f.get(url)
	if err != nil {
		return remoteData{}, err
	}
	// Parse Firestore document format
	var doc struct {
		Fields struct {
			UnlockedLevel struct{ IntegerValue string } `json:"unlockedLevel"`
			Coins         struct{ IntegerValue string } `json:"coins"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return remoteData{}, err
	}
	var rd remoteData
	fmt.Sscan(doc.Fields.UnlockedLevel.IntegerValue, &rd.UnlockedLevel)
	fmt.Sscan(doc.Fields.Coins.IntegerValue, &rd.Coins)
	return rd, nil
}

// PushSave uploads current progress to Firestore.
func (f *FirebaseClient) PushSave() error {
	if !f.ready || f.uid == "" {
		return nil
	}
	url := fmt.Sprintf("%s/users/%s?updateMask.fieldPaths=unlockedLevel&updateMask.fieldPaths=coins", fbStoreBase, f.uid)
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"unlockedLevel": map[string]interface{}{"integerValue": fmt.Sprintf("%d", progress.UnlockedLevel)},
			"coins":         map[string]interface{}{"integerValue": fmt.Sprintf("%d", progress.Coins)},
		},
	}
	body, _ := json.Marshal(payload)
	_, err := f.patch(url, body)
	return err
}

// ── Analytics ─────────────────────────────────────────────────────

// LogEvent sends an analytics event to Firebase Analytics (via Measurement Protocol).
func (f *FirebaseClient) LogEvent(name string, params map[string]interface{}) {
	if f.offline {
		return
	}
	// Fire-and-forget
	go func() {
		payload := map[string]interface{}{
			"client_id": f.uid,
			"events": []map[string]interface{}{
				{"name": name, "params": params},
			},
		}
		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=G-PLACEHOLDER&api_secret=%s", fbAPIKey)
		_, _ = f.post(url, body) // best-effort
	}()
}

func LogLevelStart(n int)  { fb.LogEvent("level_start", map[string]interface{}{"level": n}) }
func LogLevelWin(n, score int) {
	fb.LogEvent("level_complete", map[string]interface{}{"level": n, "score": score})
}
func LogPurchase(item string, coins int) {
	fb.LogEvent("spend_virtual_currency", map[string]interface{}{
		"item_name": item, "value": coins, "virtual_currency_name": "coins",
	})
}

// ── Leaderboard ───────────────────────────────────────────────────

type LeaderboardEntry struct {
	Rank  int
	Name  string
	Score int
}

// GetLeaderboard returns top 10 scores (stubbed until real API key).
func GetLeaderboard() []LeaderboardEntry {
	return []LeaderboardEntry{
		{1, "CrystalWizard", 98420},
		{2, "RuneMaster99", 87310},
		{3, "GemCrusher", 75880},
		{4, "ArcaneBlast", 64200},
		{5, "You", progress.BestScore},
	}
}

// ── HTTP helpers ──────────────────────────────────────────────────

func (f *FirebaseClient) post(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if f.idToken != "" {
		req.Header.Set("Authorization", "Bearer "+f.idToken)
	}
	resp, err := f.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (f *FirebaseClient) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if f.idToken != "" {
		req.Header.Set("Authorization", "Bearer "+f.idToken)
	}
	resp, err := f.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (f *FirebaseClient) patch(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if f.idToken != "" {
		req.Header.Set("Authorization", "Bearer "+f.idToken)
	}
	resp, err := f.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
