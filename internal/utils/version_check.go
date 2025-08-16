package utils
import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)
type VersionUtils struct{}
type ParsedVersion struct {
	Epoch   int
	Release []int
	Phase   VersionPhase
}
type VersionPhase struct {
	Weight int
	Number int
}
var versionPattern = regexp.MustCompile(`(?i)^v?(?:(?P<epoch>[0-9]+)!)?(?P<release>[0-9]+(?:\.[0-9]+)*)(?P<pre>[-._]?(?P<pre_tag>a|b|rc|alpha|beta|pre|preview)[-._]?(?P<pre_num>[0-9]+)?)?(?P<post>(?:[-._]?(?P<post_tag>post|rev|r)[-._]?(?P<post_num>[0-9]+)?))?(?P<dev>[-._]?(?P<dev_tag>dev)[-._]?(?P<dev_num>[0-9]+)?)?(?:\+(?P<local>[a-z0-9]+(?:[-._][a-z0-9]+)*))?$`)
var phaseWeights = map[string]int{
	"dev":   -4,
	"a":     -3,  
	"b":     -2,  
	"rc":    -1,  
	"final": 0,   
	"post":  1,   
}
var tagAliases = map[string]string{
	"alpha":   "a",
	"beta":    "b",
	"pre":     "rc",
	"preview": "rc",
	"rev":     "post",
	"r":       "post",
}
func (vu *VersionUtils) ParseVersion(versionStr string) (*ParsedVersion, error) {
	match := versionPattern.FindStringSubmatch(versionStr)
	if match == nil {
		return nil, fmt.Errorf("无效的版本格式: %s", versionStr)
	}
	result := make(map[string]string)
	names := versionPattern.SubexpNames()
	for i, name := range names {
		if i > 0 && name != "" && i < len(match) {
			result[name] = match[i]
		}
	}
	epoch := 0
	if epochStr := result["epoch"]; epochStr != "" {
		var err error
		epoch, err = strconv.Atoi(epochStr)
		if err != nil {
			return nil, fmt.Errorf("解析纪元失败: %v", err)
		}
	}
	releaseStr := result["release"]
	if releaseStr == "" {
		return nil, fmt.Errorf("缺少发布版本号")
	}
	releaseParts := strings.Split(releaseStr, ".")
	release := make([]int, len(releaseParts))
	for i, part := range releaseParts {
		var err error
		release[i], err = strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("解析发布版本号失败: %v", err)
		}
	}
	var phase VersionPhase
	if devTag := result["dev_tag"]; devTag != "" {
		phase.Weight = phaseWeights["dev"]
		if devNum := result["dev_num"]; devNum != "" {
			phase.Number, _ = strconv.Atoi(devNum)
		}
	} else if preTag := result["pre_tag"]; preTag != "" {
		tag := strings.ToLower(preTag)
		normalizedTag := tagAliases[tag]
		if normalizedTag == "" {
			normalizedTag = tag
		}
		phase.Weight = phaseWeights[normalizedTag]
		if preNum := result["pre_num"]; preNum != "" {
			phase.Number, _ = strconv.Atoi(preNum)
		}
	} else if postTag := result["post_tag"]; postTag != "" {
		phase.Weight = phaseWeights["post"]
		if postNum := result["post_num"]; postNum != "" {
			phase.Number, _ = strconv.Atoi(postNum)
		}
	} else {
		phase.Weight = phaseWeights["final"]
		phase.Number = 0
	}
	return &ParsedVersion{
		Epoch:   epoch,
		Release: release,
		Phase:   phase,
	}, nil
}
func (vu *VersionUtils) normalizeReleases(a, b []int) ([]int, []int) {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	aPadded := make([]int, maxLen)
	copy(aPadded, a)
	bPadded := make([]int, maxLen)
	copy(bPadded, b)
	return aPadded, bPadded
}
func (vu *VersionUtils) CompareVersions(a, b string) (int, error) {
	keyA, err := vu.ParseVersion(a)
	if err != nil {
		return 0, fmt.Errorf("解析版本 %s 失败: %v", a, err)
	}
	keyB, err := vu.ParseVersion(b)
	if err != nil {
		return 0, fmt.Errorf("解析版本 %s 失败: %v", b, err)
	}
	if keyA.Epoch > keyB.Epoch {
		return 1, nil
	}
	if keyA.Epoch < keyB.Epoch {
		return -1, nil
	}
	releaseA, releaseB := vu.normalizeReleases(keyA.Release, keyB.Release)
	for i := 0; i < len(releaseA); i++ {
		if releaseA[i] > releaseB[i] {
			return 1, nil
		}
		if releaseA[i] < releaseB[i] {
			return -1, nil
		}
	}
	if keyA.Phase.Weight > keyB.Phase.Weight {
		return 1, nil
	}
	if keyA.Phase.Weight < keyB.Phase.Weight {
		return -1, nil
	}
	if keyA.Phase.Number > keyB.Phase.Number {
		return 1, nil
	}
	if keyA.Phase.Number < keyB.Phase.Number {
		return -1, nil
	}
	return 0, nil
}
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
}
type CacheItem struct {
	Release   *GitHubRelease
	HasUpdate bool
	Error     error
	ExpiresAt time.Time
}
type UpdateChecker struct {
	vu         *VersionUtils
	client     *http.Client
	cache      map[string]*CacheItem
	mutex      sync.RWMutex
	lastCall   time.Time
	callMutex  sync.Mutex
	logger     LoggerInterface   
}
type LoggerInterface interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}
func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{
		vu: &VersionUtils{},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*CacheItem),
	}
}
func NewUpdateCheckerWithLogger(logger LoggerInterface) *UpdateChecker {
	return &UpdateChecker{
		vu: &VersionUtils{},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:  make(map[string]*CacheItem),
		logger: logger,
	}
}
func (uc *UpdateChecker) safeLog(level string, format string, args ...interface{}) {
	if uc.logger == nil {
		return
	}
	switch level {
	case "debug":
		uc.logger.Debug(format, args...)
	case "info":
		uc.logger.Info(format, args...)
	case "warn":
		uc.logger.Warn(format, args...)
	case "error":
		uc.logger.Error(format, args...)
	}
}
func (uc *UpdateChecker) CheckForUpdates(repoOwner, repoName, currentVersion string) (*GitHubRelease, bool, error) {
	cacheKey := fmt.Sprintf("%s/%s:%s", repoOwner, repoName, currentVersion)
	uc.safeLog("debug", "检查更新: %s/%s (当前版本: %s)", repoOwner, repoName, currentVersion)
	uc.mutex.RLock()
	if item, exists := uc.cache[cacheKey]; exists && time.Now().Before(item.ExpiresAt) {
		uc.mutex.RUnlock()
		uc.safeLog("debug", "使用缓存结果: %s/%s", repoOwner, repoName)
		return item.Release, item.HasUpdate, item.Error
	}
	uc.mutex.RUnlock()
	uc.callMutex.Lock()
	since := time.Since(uc.lastCall)
	if since < 2*time.Second {
		waitTime := 2*time.Second - since
		uc.safeLog("info", "GitHub API 调用频率限制，等待 %v 后继续", waitTime)
		time.Sleep(waitTime)
	}
	uc.lastCall = time.Now()
	uc.callMutex.Unlock()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	uc.safeLog("debug", "请求 GitHub API: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		uc.safeLog("error", "创建 GitHub API 请求失败: %v (repo: %s/%s)", err, repoOwner, repoName)
		return nil, false, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "LazyTea-Mobile/1.0 (https://github.com/user/lazytea-mobile)")
	resp, err := uc.client.Do(req)
	if err != nil {
		uc.safeLog("error", "GitHub API 网络请求失败: %v (repo: %s/%s)", err, repoOwner, repoName)
		uc.cacheResult(cacheKey, nil, false, fmt.Errorf("获取发布信息失败: %v", err))
		return nil, false, fmt.Errorf("获取发布信息失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errorMsg string
		switch resp.StatusCode {
		case 403:
			rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
			rateLimitReset := resp.Header.Get("X-RateLimit-Reset")
			if rateLimitRemaining == "0" {
				resetTime := "未知"
				if rateLimitReset != "" {
					if resetTimestamp, parseErr := strconv.ParseInt(rateLimitReset, 10, 64); parseErr == nil {
						resetTime = time.Unix(resetTimestamp, 0).Format("15:04:05")
					}
				}
				errorMsg = fmt.Sprintf("GitHub API 速率限制 (剩余: %s, 重置时间: %s)", rateLimitRemaining, resetTime)
				uc.safeLog("error", "GitHub API 速率限制: repo=%s/%s, remaining=%s, reset=%s", repoOwner, repoName, rateLimitRemaining, resetTime)
			} else {
				errorMsg = fmt.Sprintf("GitHub API 访问被拒绝 (状态码: %d)", resp.StatusCode)
				uc.safeLog("error", "GitHub API 访问被拒绝: repo=%s/%s, status=%d", repoOwner, repoName, resp.StatusCode)
			}
		case 404:
			errorMsg = fmt.Sprintf("GitHub 仓库或发布信息未找到 (状态码: %d)", resp.StatusCode)
			uc.safeLog("warn", "GitHub 仓库或发布信息未找到: repo=%s/%s", repoOwner, repoName)
		default:
			errorMsg = fmt.Sprintf("GitHub API 请求失败 (状态码: %d)", resp.StatusCode)
			uc.safeLog("error", "GitHub API 请求失败: repo=%s/%s, status=%d", repoOwner, repoName, resp.StatusCode)
		}
		err := fmt.Errorf(errorMsg)
		uc.cacheResult(cacheKey, nil, false, err)
		return nil, false, err
	}
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		uc.safeLog("error", "解析 GitHub API 响应失败: %v (repo: %s/%s)", err, repoOwner, repoName)
		err := fmt.Errorf("解析发布信息失败: %v", err)
		uc.cacheResult(cacheKey, nil, false, err)
		return nil, false, err
	}
	uc.safeLog("debug", "成功获取发布信息: repo=%s/%s, latest=%s, draft=%v, prerelease=%v", 
		repoOwner, repoName, release.TagName, release.Draft, release.Prerelease)
	if release.Draft || release.Prerelease {
		uc.safeLog("debug", "跳过草稿或预发布版本: repo=%s/%s, version=%s", repoOwner, repoName, release.TagName)
		uc.cacheResult(cacheKey, &release, false, nil)
		return &release, false, nil
	}
	result, err := uc.vu.CompareVersions(release.TagName, currentVersion)
	if err != nil {
		uc.safeLog("error", "版本比较失败: repo=%s/%s, latest=%s, current=%s, error=%v", 
			repoOwner, repoName, release.TagName, currentVersion, err)
		err := fmt.Errorf("版本比较失败: %v", err)
		uc.cacheResult(cacheKey, &release, false, err)
		return &release, false, err
	}
	hasUpdate := result > 0
	uc.safeLog("debug", "版本比较结果: repo=%s/%s, latest=%s, current=%s, hasUpdate=%v", 
		repoOwner, repoName, release.TagName, currentVersion, hasUpdate)
	uc.cacheResult(cacheKey, &release, hasUpdate, nil)
	return &release, hasUpdate, nil
}
func (uc *UpdateChecker) cacheResult(cacheKey string, release *GitHubRelease, hasUpdate bool, err error) {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	expiresAt := time.Now().Add(12 * time.Hour)
	uc.cache[cacheKey] = &CacheItem{
		Release:   release,
		HasUpdate: hasUpdate,
		Error:     err,
		ExpiresAt: expiresAt,
	}
}
func (uc *UpdateChecker) CleanExpiredCache() {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	now := time.Now()
	for key, item := range uc.cache {
		if now.After(item.ExpiresAt) {
			delete(uc.cache, key)
		}
	}
}
func (uc *UpdateChecker) IsVersionNewer(versionA, versionB string) (bool, error) {
	result, err := uc.vu.CompareVersions(versionA, versionB)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}