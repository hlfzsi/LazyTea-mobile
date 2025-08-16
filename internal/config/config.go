package config
import (
	"encoding/json"
	"log"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)
type Config struct {
	Database DatabaseConfig `json:"database"`
	Network  NetworkConfig  `json:"network"`
}
type DatabaseConfig struct {
	Path string `json:"path"`  
}
type NetworkConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Token        string `json:"token"`
	AutoConnect  bool   `json:"auto_connect"`
	RememberAuth bool   `json:"remember_auth"`
}
var (
	globalConfig *Config
	appInstance fyne.App
)
func SetApp(app fyne.App) {
	appInstance = app
}
func Load() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}
	configURI, err := getConfigURI()
	if err != nil {
		return nil, err
	}
	canRead, err := storage.CanRead(configURI)
	if err != nil || !canRead {
		globalConfig = getDefaultConfig()
		saveErr := Save(globalConfig)
		if saveErr != nil {
			log.Printf("警告：无法保存默认配置文件：%v", saveErr)
		}
		return globalConfig, nil  
	}
	reader, err := storage.Reader(configURI)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var config Config
	if err := json.NewDecoder(reader).Decode(&config); err != nil {
		return nil, err
	}
	globalConfig = &config
	return globalConfig, nil
}
func Save(config *Config) error {
	configURI, err := getConfigURI()
	if err != nil {
		return err
	}
	writer, err := storage.Writer(configURI)
	if err != nil {
		return err
	}
	defer writer.Close()
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
func GetGlobal() *Config {
	if globalConfig == nil {
		var err error
		globalConfig, err = Load()
		if err != nil {
			log.Printf("错误：加载配置失败：%v。将使用默认配置。", err)
			globalConfig = getDefaultConfig()
		}
	}
	return globalConfig
}
func checkAppInstance() error {
	if appInstance == nil {
		log.Panic("错误：config包的appInstance未设置。请在程序启动时调用 config.SetApp()。")
	}
	return nil
}
func getConfigURI() (fyne.URI, error) {
	if err := checkAppInstance(); err != nil {
		return nil, err
	}
	rootURI := appInstance.Storage().RootURI()
	return storage.Child(rootURI, "config.json")
}
func GetDatabaseURI() (fyne.URI, error) {
	if err := checkAppInstance(); err != nil {
		return nil, err
	}
	rootURI := appInstance.Storage().RootURI()
	return storage.Child(rootURI, "data.db")
}
func getDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path: func() string {
				uri, err := GetDatabaseURI()
				if err != nil {
					log.Printf("警告:无法获取数据库URI：%v", err)
					return "err"
				}
				return uri.String()
			}(),
		},
		Network: NetworkConfig{
			Host:         "127.0.0.1",
			Port:         8080,
			Token:        "疯狂星期四V我50",
			AutoConnect:  false,
			RememberAuth: true,
		},
	}
}
