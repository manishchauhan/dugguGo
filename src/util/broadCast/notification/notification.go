package notification

import (
	"sync"
)

type NotifySystem struct {
	notifications map[string]interface{}
	mutex         sync.RWMutex
}

var notifySystemInstance *NotifySystem
var notifySystemOnce sync.Once

// GetNotifySystemInstance returns the singleton instance of NotifySystem.
func GetNotifySystemInstance() *NotifySystem {
	notifySystemOnce.Do(func() {
		notifySystemInstance = &NotifySystem{
			notifications: make(map[string]interface{}),
		}
	})
	return notifySystemInstance
}

func (ns *NotifySystem) AddNotification(keyname string, data interface{}) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	ns.notifications[keyname] = data
}

func (ns *NotifySystem) ReceiveNotification(keyname string, notificationFunction func(interface{})) {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	data, exists := ns.notifications[keyname]
	if exists {
		notificationFunction(data)
	}
}
