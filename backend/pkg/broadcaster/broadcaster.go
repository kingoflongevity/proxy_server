package broadcaster

import (
	"sync"
)

// LogBroadcastFunc 日志广播函数类型
type LogBroadcastFunc func(level, message, source string)

// NodeBroadcastFunc 节点广播函数类型
type NodeBroadcastFunc func(nodeId, status string, latency int)

// SubscriptionBroadcastFunc 订阅广播函数类型
type SubscriptionBroadcastFunc func(subscriptionId string, nodeCount int, status string)

var (
	logBroadcastFunc          LogBroadcastFunc
	nodeBroadcastFunc         NodeBroadcastFunc
	subscriptionBroadcastFunc SubscriptionBroadcastFunc
	broadcasterMutex          sync.RWMutex
)

// RegisterLogBroadcast 注册日志广播函数
func RegisterLogBroadcast(fn LogBroadcastFunc) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()
	logBroadcastFunc = fn
}

// RegisterNodeBroadcast 注册节点广播函数
func RegisterNodeBroadcast(fn NodeBroadcastFunc) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()
	nodeBroadcastFunc = fn
}

// RegisterSubscriptionBroadcast 注册订阅广播函数
func RegisterSubscriptionBroadcast(fn SubscriptionBroadcastFunc) {
	broadcasterMutex.Lock()
	defer broadcasterMutex.Unlock()
	subscriptionBroadcastFunc = fn
}

// BroadcastLog 广播日志
func BroadcastLog(level, message, source string) {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()
	if logBroadcastFunc != nil {
		logBroadcastFunc(level, message, source)
	}
}

// BroadcastNode 广播节点更新
func BroadcastNode(nodeId, status string, latency int) {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()
	if nodeBroadcastFunc != nil {
		nodeBroadcastFunc(nodeId, status, latency)
	}
}

// BroadcastSubscription 广播订阅更新
func BroadcastSubscription(subscriptionId string, nodeCount int, status string) {
	broadcasterMutex.RLock()
	defer broadcasterMutex.RUnlock()
	if subscriptionBroadcastFunc != nil {
		subscriptionBroadcastFunc(subscriptionId, nodeCount, status)
	}
}
