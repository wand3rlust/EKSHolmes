package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type jwtParts struct {
	Header  map[string]any `json:"header"`
	Payload map[string]any `json:"payload"`
	Raw     string         `json:"raw"`
}

func readToken(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(string(b))
	if tok == "" {
		return "", errors.New("[X] Empty token")
	}
	return tok, nil
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

func parseJWT(token string) (*jwtParts, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("[X] Invalid JWT structure")
	}

	hRaw, err := decodeSegment(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	pRaw, err := decodeSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var header map[string]any
	var payload map[string]any
	if err := json.Unmarshal(hRaw, &header); err != nil {
		return nil, fmt.Errorf("unmarshal header: %w", err)
	}
	if err := json.Unmarshal(pRaw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	return &jwtParts{
		Header:  header,
		Payload: payload,
		Raw:     token,
	}, nil
}

func extractEKSInfo(payload map[string]any) (clusterID, region string, err error) {
	iss, ok := payload["iss"].(string)
	if !ok {
		return "", "", errors.New("[X] iss field not found or corrupted in JWT payload")
	}

	re := regexp.MustCompile(`https://oidc\.eks\.([^.]+)\.amazonaws\.com/id/([A-F0-9]+)`)
	matches := re.FindStringSubmatch(iss)
	if len(matches) != 3 {
		return "", "", errors.New("[X] Invalid iss format")
	}

	return matches[2], matches[1], nil
}

func enumerateEKSEndpoint(clusterID, region string) (string, error) {
	codes := []string{
		"gr5", "sk1", "uw2", "ue1", "ew1", "ap1", "se1", "ne1",
		"sk2", "sk3", "sk4", "sk5", "sk6", "sk7", "sk8", "sk9",
		"gr1", "gr2", "gr3", "gr4", "gr7", "gr6", "gr8", "gr9",
		"ue2", "ue3", "ue4", "ue5", "ue6", "ue7", "ue8", "ue9",
		"ew2", "ew3", "ew4", "ew5", "ew6", "ew7", "ew8", "ew9",
		"ap2", "ap3", "ap4", "ap5", "ap6", "ap7", "ap8", "ap9",
		"se2", "se3", "se4", "se5", "se6", "se7", "se8", "se9",
		"ne2", "ne3", "ne4", "ne5", "ne6", "ne7", "ne8", "ne9",
		"uw1", "uw3", "uw4", "uw5", "uw6", "uw7", "uw8", "uw9",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	fmt.Printf("\n[+] Enumerating EKS endpoint for cluster %s in %s...\n", colorize(clusterID, GREEN), colorize(region, GREEN))
	fmt.Printf("\n[+] Progress: ")

	for i, code := range codes {
		url := fmt.Sprintf("https://%s.%s.%s.eks.amazonaws.com", clusterID, code, region)

		// Show progress dots instead of full URLs
		if i%10 == 0 {
			fmt.Printf("[%d/%d]", i+1, len(codes))
		} else {
			fmt.Printf(".")
		}

		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		resp.Body.Close()

		if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 200 {
			fmt.Printf("\n[✓] Found valid endpoint: %s (Status: %d)\n", colorize(url, GREEN), resp.StatusCode)
			return url, nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n")
	return "", errors.New("[X] No valid endpoint found")
}

func eksAPIEnum() {
	token, err := readToken(tokenPath)
	if err != nil {
		fmt.Printf("[X] Read token error: %v\n", err)
		return
	}

	j, err := parseJWT(token)
	if err != nil {
		fmt.Printf("[X] Parse JWT error: %v\n", err)
		return
	}

	clusterID, region, err := extractEKSInfo(j.Payload)
	if err != nil {
		fmt.Printf("[X] Extract EKS info error: %v\n", err)
		return
	}

	fmt.Printf("[✓] Cluster ID: %s\n", colorize(clusterID, GREEN))
	fmt.Printf("[✓] Region: %s\n", colorize(region, GREEN))

	endpoint, err := enumerateEKSEndpoint(clusterID, region)
	if err != nil {
		fmt.Printf("[X] Endpoint discovery error: %v\n", err)
	} else {
		fmt.Printf("[✓] EKS API Server: %s\n", colorize(endpoint, GREEN))
	}
}
