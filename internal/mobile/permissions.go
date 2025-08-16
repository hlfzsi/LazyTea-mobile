package mobile
import (
	"log"
)
type PermissionType int
const (
	PermissionInternet PermissionType = iota
	PermissionNetworkState
	PermissionWakeLock
	PermissionStorage
	PermissionNotification
)
type PermissionManager struct {
	requestedPermissions map[PermissionType]bool
	grantedPermissions   map[PermissionType]bool
}
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		requestedPermissions: make(map[PermissionType]bool),
		grantedPermissions:   make(map[PermissionType]bool),
	}
}
func (pm *PermissionManager) RequestPermission(permission PermissionType) bool {
	if !IsMobile() {
		pm.grantedPermissions[permission] = true
		return true
	}
	pm.requestedPermissions[permission] = true
	switch permission {
	case PermissionInternet:
		pm.grantedPermissions[permission] = true
		log.Println("[Mobile] Internet permission granted")
		return true
	case PermissionNetworkState:
		pm.grantedPermissions[permission] = true
		log.Println("[Mobile] Network state permission granted")
		return true
	case PermissionWakeLock:
		pm.grantedPermissions[permission] = true
		log.Println("[Mobile] Wake lock permission granted")
		return true
	case PermissionStorage:
		pm.grantedPermissions[permission] = true
		log.Println("[Mobile] Storage permission granted (using app-specific directories)")
		return true
	case PermissionNotification:
		pm.grantedPermissions[permission] = true
		log.Println("[Mobile] Notification permission granted")
		return true
	default:
		log.Printf("[Mobile] Unknown permission type: %v", permission)
		return false
	}
}
func (pm *PermissionManager) HasPermission(permission PermissionType) bool {
	granted, exists := pm.grantedPermissions[permission]
	return exists && granted
}
func (pm *PermissionManager) RequestAllPermissions() {
	log.Println("[Mobile] Requesting all necessary permissions...")
	permissions := []PermissionType{
		PermissionInternet,
		PermissionNetworkState,
		PermissionWakeLock,
		PermissionStorage,
	}
	for _, permission := range permissions {
		pm.RequestPermission(permission)
	}
}
func (pm *PermissionManager) GetPermissionStatus() map[string]bool {
	status := make(map[string]bool)
	status["internet"] = pm.HasPermission(PermissionInternet)
	status["network_state"] = pm.HasPermission(PermissionNetworkState)
	status["wake_lock"] = pm.HasPermission(PermissionWakeLock)
	status["storage"] = pm.HasPermission(PermissionStorage)
	status["notification"] = pm.HasPermission(PermissionNotification)
	return status
}
type NetworkPermissions struct {
	*PermissionManager
}
func NewNetworkPermissions() *NetworkPermissions {
	return &NetworkPermissions{
		PermissionManager: NewPermissionManager(),
	}
}
func (np *NetworkPermissions) EnsureNetworkPermissions() bool {
	success := true
	if !np.RequestPermission(PermissionInternet) {
		log.Println("[Mobile] Failed to get internet permission")
		success = false
	}
	if !np.RequestPermission(PermissionNetworkState) {
		log.Println("[Mobile] Failed to get network state permission")
		success = false
	}
	return success
}
type StoragePermissions struct {
	*PermissionManager
}
func NewStoragePermissions() *StoragePermissions {
	return &StoragePermissions{
		PermissionManager: NewPermissionManager(),
	}
}
func (sp *StoragePermissions) EnsureStoragePermissions() bool {
	return sp.RequestPermission(PermissionStorage)
}
func CheckAndRequestPermissions() *PermissionManager {
	pm := NewPermissionManager()
	if IsMobile() {
		log.Println("[Mobile] Running on mobile device, checking permissions...")
		pm.RequestAllPermissions()
	} else {
		log.Println("[Mobile] Running on desktop, granting all permissions...")
		permissions := []PermissionType{
			PermissionInternet,
			PermissionNetworkState,
			PermissionWakeLock,
			PermissionStorage,
			PermissionNotification,
		}
		for _, permission := range permissions {
			pm.grantedPermissions[permission] = true
		}
	}
	status := pm.GetPermissionStatus()
	for perm, granted := range status {
		if granted {
			log.Printf("[Mobile] Permission %s: GRANTED", perm)
		} else {
			log.Printf("[Mobile] Permission %s: DENIED", perm)
		}
	}
	return pm
}
