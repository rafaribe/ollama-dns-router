package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Settings struct {
	Instances []string `mapstructure:"instances"`
	Provider  struct {
		Pihole struct {
			Hostname string `mapstructure:"hostname"`
			APIKey   string `mapstructure:"apiKey"`
		} `mapstructure:"pihole"`
	} `mapstructure:"provider"`
	OllamaRecord string `mapstructure:"ollamaRecord"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type InstanceChecker struct {
	settings *Settings
	logger   *zap.Logger
	client   *http.Client
}

func NewInstanceChecker(settings *Settings, logger *zap.Logger) *InstanceChecker {
	return &InstanceChecker{
		settings: settings,
		logger:   logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

}

func (ic *InstanceChecker) checkAvailability(url string) (bool, error) {
	resp, err := ic.client.Get(url)
	if err != nil {
		ic.logger.Error("health check failed", zap.Error(err), zap.String("url", url))
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ic.logger.Info("instance health check failed", zap.Int("status", resp.StatusCode), zap.String("url", url))
		return false, nil
	}

	return true, nil
}

type DNSManager struct {
	logger   *zap.Logger
	settings *Settings
}

func NewDNSManager(logger *zap.Logger, settings *Settings) *DNSManager {
	return &DNSManager{
		logger:   logger,
		settings: settings,
	}
}

func (dm *DNSManager) createDNSRecord(instanceURL, piholeHost, apiKey, ollamaRecord string) error {
	parsedURL, err := url.Parse(instanceURL)
	if err != nil {
		dm.logger.Error("invalid instance URL", zap.Error(err))
		return err
	}

	host := parsedURL.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			dm.logger.Error("failed to resolve IP", zap.Error(err))
			return err
		}
		ip = ips[0]
	}

	baseURL := fmt.Sprintf("http://%s/admin/api.php", piholeHost)

	// Delete existing record
	deleteURL := fmt.Sprintf("%s?customdns&action=delete&domain=%s&ip=%s&auth=%s",
		baseURL, ollamaRecord, ip.String(), apiKey)

	deleteReq, err := http.NewRequest(http.MethodPost, deleteURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		dm.logger.Error("failed to delete existing DNS record", zap.Error(err))
		return err
	}
	resp.Body.Close()

	// Create new record
	createURL := fmt.Sprintf("%s?customdns&action=add&domain=%s&ip=%s&auth=%s",
		baseURL, ollamaRecord, ip.String(), apiKey)

	createReq, err := http.NewRequest(http.MethodPost, createURL, nil)
	if err != nil {
		return err
	}

	resp, err = http.DefaultClient.Do(createReq)
	if err != nil {
		dm.logger.Error("failed to create DNS record", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	dm.logger.Info("DNS record created successfully",
		zap.String("ip", ip.String()),
		zap.String("domain", ollamaRecord))
	return nil
}
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	settings, err := loadSettings()
	if err != nil {
		logger.Fatal("failed to load settings", zap.Error(err))
	}

	checker := NewInstanceChecker(settings, logger)
	dnsManager := NewDNSManager(logger, settings)

	var availableInstance string
	for _, instance := range settings.Instances {
		if available, err := checker.checkAvailability(instance); err == nil && available {
			availableInstance = instance
			break
		}
	}

	if availableInstance == "" {
		logger.Fatal("no available instances found")
	}

	if err := dnsManager.createDNSRecord(
		availableInstance,
		settings.Provider.Pihole.Hostname,
		settings.Provider.Pihole.APIKey,
		settings.OllamaRecord,
	); err != nil {
		logger.Fatal("failed to create DNS record", zap.Error(err))
	}

	logger.Info("Exiting")
}

func loadSettings() (*Settings, error) {
	var settings Settings

	// Load from environment variables
	if envURLs := os.Getenv("INSTANCE_URLS"); envURLs != "" {
		settings.Instances = strings.Split(envURLs, ",")
	}

	if ollamaRecord := os.Getenv("OLLAMA_RECORD"); ollamaRecord != "" {
		settings.OllamaRecord = ollamaRecord
	}

	if piholeHostname := os.Getenv("PIHOLE_HOSTNAME"); piholeHostname != "" {
		settings.Provider.Pihole.Hostname = piholeHostname
	}

	if piholeAPIKey := os.Getenv("PIHOLE_API_KEY"); piholeAPIKey != "" {
		settings.Provider.Pihole.APIKey = piholeAPIKey
	}

	return &settings, nil
}
