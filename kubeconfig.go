package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const (
	caCertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

type KubeconfigData struct {
	ServiceAccount string
	Namespace      string
	Token          string
	CACert         string
	Server         string
	ClusterName    string
}

func extractServiceAccountFromJWT(payload map[string]any) (string, string, error) {
	sub, ok := payload["sub"].(string)
	if !ok {
		return "", "", fmt.Errorf("[X] sub field not found in JWT payload")
	}

	// Extract service account and namespace from sub field
	// Format: system:serviceaccount:namespace:serviceaccount-name
	parts := strings.Split(sub, ":")
	if len(parts) != 4 || parts[0] != "system" || parts[1] != "serviceaccount" {
		return "", "", fmt.Errorf("[X] Invalid sub format: %s", sub)
	}

	namespace := parts[2]
	serviceAccount := parts[3]

	return serviceAccount, namespace, nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("[X] Failed to read %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func getCAFromFile() (string, error) {
	caCert, err := readFile(caCertPath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(caCert)), nil
}

func getClusterNameFromEKS(server string) string {
	// Extract cluster ID from server URL for cluster name
	// Example: https://ABC123.gr7.us-west-2.eks.amazonaws.com -> ABC123
	parts := strings.Split(server, ".")
	if len(parts) > 0 {
		serverPart := strings.TrimPrefix(parts[0], "https://")
		return serverPart
	}
	return "eks-cluster"
}

func generateKubeconfig(data KubeconfigData) error {
	filename := fmt.Sprintf("/tmp/%s-%s-kubeconfig.yaml", data.ServiceAccount, data.Namespace)

	contextName := fmt.Sprintf("%s-context", data.ServiceAccount)

	kubeconfigContent := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: %s
  cluster:
    server: %s
    certificate-authority-data: %s
contexts:
- name: %s
    cluster: %s
    user: %s
    namespace: %s
current-context: %s
users:
- name: %s
  user:
    token: %s
`, data.ClusterName, data.Server, data.CACert, contextName, data.ClusterName,
		data.ServiceAccount, data.Namespace, contextName, data.ServiceAccount, data.Token)

	err := os.WriteFile(filename, []byte(kubeconfigContent), 0600)
	if err != nil {
		return fmt.Errorf("[X] Failed to write kubeconfig file: %w", err)
	}
	fmt.Printf("[✓] Kubeconfig generated successfully: %s\n", colorize(filename, GREEN))
	return nil
}

func kubeconfigGenerator() {
	token, err := readToken(tokenPath)
	if err != nil {
		fmt.Printf("[X] Error reading token: %v\n", err)
		return
	}

	jwt, err := parseJWT(token)
	if err != nil {
		fmt.Printf("[X] Error parsing JWT: %v\n", err)
		return
	}

	// Extract service account and namespace from JWT
	serviceAccount, namespace, err := extractServiceAccountFromJWT(jwt.Payload)
	if err != nil {
		fmt.Printf("[X] Error extracting serviceaccount info: %v\n", err)
		return
	}

	fmt.Printf("[✓] Service Account: %s\n", serviceAccount)
	fmt.Printf("[✓] Namespace: %s\n", namespace)

	// Get CA certificate
	caCert, err := getCAFromFile()
	if err != nil {
		fmt.Printf("[X] Error reading CA certificate: %v\n", err)
		return
	}

	// Discover EKS endpoint
	clusterID, region, err := extractEKSInfo(jwt.Payload)
	if err != nil {
		fmt.Printf("[X] Error extracting EKS info: %v\n", err)
		return
	}

	server, err := enumerateEKSEndpoint(clusterID, region)
	if err != nil {
		fmt.Printf("[X] Error discovering EKS endpoint: %v\n", err)
		return
	}

	clusterName := getClusterNameFromEKS(server)

	// Generate kubeconfig
	kubeconfigData := KubeconfigData{
		ServiceAccount: serviceAccount,
		Namespace:      namespace,
		Token:          token,
		CACert:         caCert,
		Server:         server,
		ClusterName:    clusterName,
	}

	err = generateKubeconfig(kubeconfigData)
	if err != nil {
		fmt.Printf("[X] Error generating kubeconfig: %v\n", err)
		return
	}
}
