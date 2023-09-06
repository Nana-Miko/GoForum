package throttle

import (
	sqlite "GoForum/go/db"
	"GoForum/go/util"
)

// NewThrottler 实例化节流器
func NewThrottler(ThrottleTime int64) Throttler {
	rt := restrictThrottler{
		ThrottleTime:    ThrottleTime,
		LastResponseMap: make(map[string]map[int]int64),
		PassFunc: func() {

		},
		RestrictFunc: func() {

		},
	}
	return rt
}

// Throttler 节流器
type Throttler interface {
	// ThrottleVerify 节流验证
	ThrottleVerify(user sqlite.User, id int) bool
	// SetPassFunc 设置通行方法
	SetPassFunc(f func())
	// SetRestrictFunc 设置限制方法
	SetRestrictFunc(f func())
}

// restrictThrottler 限制节流器
type restrictThrottler struct {
	ThrottleTime    int64
	LastResponseMap map[string]map[int]int64
	PassFunc        func()
	RestrictFunc    func()
}

func (rt restrictThrottler) SetPassFunc(f func()) {
	rt.PassFunc = f
}

func (rt restrictThrottler) SetRestrictFunc(f func()) {
	rt.RestrictFunc = f
}

func (rt restrictThrottler) ThrottleVerify(user sqlite.User, id int) bool {
	_, exists := rt.LastResponseMap[user.Email]
	if !exists {
		rt.LastResponseMap[user.Email] = make(map[int]int64)
	}
	_, exists = rt.LastResponseMap[user.Email][id]
	if !exists {
		rt.LastResponseMap[user.Email][id] = 0
	}
	timestamp := rt.LastResponseMap[user.Email][id]
	nowTimestamp := util.CurrentTimeStampMilli()
	// 如果节流时间到
	if nowTimestamp-timestamp >= rt.ThrottleTime {
		rt.PassFunc()
		rt.LastResponseMap[user.Email][id] = nowTimestamp
		return true
	} else {
		rt.RestrictFunc()
		return false
	}
}

// switchThrottler 开关节流器
type switchThrottler struct {
	ThrottleTime    int64
	LastResponseMap map[string]map[int]int64
	PassFunc        func()
	RestrictFunc    func()
}

func (st switchThrottler) SetPassFunc(f func()) {
	st.PassFunc = f
}

func (st switchThrottler) SetRestrictFunc(f func()) {
	st.RestrictFunc = f
}
