//go:build wasm

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"syscall/js"

	"readwillbe/types"

	"github.com/pkg/errors"
)

var baseURL = ""

func getToken() string {
	token := js.Global().Get("localStorage").Call("getItem", "token")
	if token.IsNull() || token.IsUndefined() {
		return ""
	}
	return token.String()
}

func SetToken(token string) {
	js.Global().Get("localStorage").Call("setItem", "token", token)
}

func ClearToken() {
	js.Global().Get("localStorage").Call("removeItem", "token")
}

func IsAuthenticated() bool {
	token := getToken()
	return token != "" && token != "null" && token != "undefined"
}

func makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "marshaling request body")
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	req.Header.Set("Content-Type", "application/json")
	if token := getToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	return resp, nil
}

type AuthResponse struct {
	Token     string     `json:"token"`
	ExpiresAt string     `json:"expires_at"`
	User      types.User `json:"user"`
}

func SignIn(email, password string) (*AuthResponse, error) {
	resp, err := makeRequest("POST", "/api/auth/sign-in", map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sign in request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("sign in failed with status %d", resp.StatusCode)
	}

	var result AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding sign in response")
	}
	return &result, nil
}

func SignUp(name, email, password string) (*AuthResponse, error) {
	resp, err := makeRequest("POST", "/api/auth/sign-up", map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sign up request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("sign up failed with status %d", resp.StatusCode)
	}

	var result AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding sign up response")
	}
	return &result, nil
}

type DashboardData struct {
	TodayReadings   []types.Reading `json:"today_readings"`
	OverdueReadings []types.Reading `json:"overdue_readings"`
}

func GetDashboard() (*DashboardData, error) {
	resp, err := makeRequest("GET", "/api/dashboard", nil)
	if err != nil {
		return nil, errors.Wrap(err, "dashboard request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("dashboard request failed with status %d", resp.StatusCode)
	}

	var result DashboardData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding dashboard response")
	}
	return &result, nil
}

func GetCurrentUser() (*types.User, error) {
	resp, err := makeRequest("GET", "/api/user", nil)
	if err != nil {
		return nil, errors.Wrap(err, "user request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("user request failed with status %d", resp.StatusCode)
	}

	var result types.User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding user response")
	}
	return &result, nil
}

func GetPlans() ([]types.Plan, error) {
	resp, err := makeRequest("GET", "/api/plans", nil)
	if err != nil {
		return nil, errors.Wrap(err, "plans request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("plans request failed with status %d", resp.StatusCode)
	}

	var result []types.Plan
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding plans response")
	}
	return result, nil
}

func GetPlan(id uint) (*types.Plan, error) {
	resp, err := makeRequest("GET", "/api/plans/"+strconv.FormatUint(uint64(id), 10), nil)
	if err != nil {
		return nil, errors.Wrap(err, "plan request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("plan request failed with status %d", resp.StatusCode)
	}

	var result types.Plan
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding plan response")
	}
	return &result, nil
}

func DeletePlan(id uint) error {
	resp, err := makeRequest("DELETE", "/api/plans/"+strconv.FormatUint(uint64(id), 10), nil)
	if err != nil {
		return errors.Wrap(err, "delete plan request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete plan failed with status %d", resp.StatusCode)
	}
	return nil
}

func RenamePlan(id uint, title string) (*types.Plan, error) {
	resp, err := makeRequest("POST", "/api/plans/"+strconv.FormatUint(uint64(id), 10)+"/rename", map[string]string{
		"title": title,
	})
	if err != nil {
		return nil, errors.Wrap(err, "rename plan request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("rename plan failed with status %d", resp.StatusCode)
	}

	var result types.Plan
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding rename plan response")
	}
	return &result, nil
}

func CompleteReading(id uint) (*types.Reading, error) {
	resp, err := makeRequest("POST", "/api/readings/"+strconv.FormatUint(uint64(id), 10)+"/complete", nil)
	if err != nil {
		return nil, errors.Wrap(err, "complete reading request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("complete reading failed with status %d", resp.StatusCode)
	}

	var result types.Reading
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding complete reading response")
	}
	return &result, nil
}

func UncompleteReading(id uint) (*types.Reading, error) {
	resp, err := makeRequest("POST", "/api/readings/"+strconv.FormatUint(uint64(id), 10)+"/uncomplete", nil)
	if err != nil {
		return nil, errors.Wrap(err, "uncomplete reading request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("uncomplete reading failed with status %d", resp.StatusCode)
	}

	var result types.Reading
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding uncomplete reading response")
	}
	return &result, nil
}

func DeleteReading(id uint) error {
	resp, err := makeRequest("DELETE", "/api/readings/"+strconv.FormatUint(uint64(id), 10), nil)
	if err != nil {
		return errors.Wrap(err, "delete reading request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("delete reading failed with status %d", resp.StatusCode)
	}
	return nil
}

func GetHistory() ([]types.Reading, error) {
	resp, err := makeRequest("GET", "/api/history", nil)
	if err != nil {
		return nil, errors.Wrap(err, "history request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("history request failed with status %d", resp.StatusCode)
	}

	var result []types.Reading
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding history response")
	}
	return result, nil
}

func UpdateSettings(notificationsEnabled bool, notificationTime string) (*types.User, error) {
	resp, err := makeRequest("PUT", "/api/account/settings", map[string]interface{}{
		"notifications_enabled": notificationsEnabled,
		"notification_time":     notificationTime,
	})
	if err != nil {
		return nil, errors.Wrap(err, "update settings request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("update settings failed with status %d", resp.StatusCode)
	}

	var result types.User
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decoding update settings response")
	}
	return &result, nil
}
