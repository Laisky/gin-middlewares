package middlewares

import (
	"context"
	"sync"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"
)

// LockableMw create a gin middleware that add a lock to gin context
func LockableMw() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lock := &sync.RWMutex{}
		ctx.Set(string(CtxKeyLock), lock)
		ctx.Next()
	}
}

// GetLockFromCtx get lock from gin context
func GetLockFromCtx(ctx context.Context) (*sync.RWMutex, error) {
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	gctx, ok := GetGinCtxFromStdCtx(ctx)
	if !ok {
		return nil, errors.Errorf("gin context not found in ctx")
	}

	if locki, exists := gctx.Get(string(CtxKeyLock)); exists {
		lock, ok := locki.(*sync.RWMutex)
		if !ok {
			return nil, errors.Errorf("%q in ctx is not a *sync.RWMutex", string(CtxKeyLock))
		}

		return lock, nil
	}

	return nil, errors.Errorf("%q not found in ctx", string(CtxKeyLock))
}

// CtxLock lock the lock in gin context
func CtxLock(ctx context.Context) error {
	lock, err := GetLockFromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "get lock from ctx")
	}

	lock.Lock()
	return nil
}

// CtxUnlock unlock the lock in gin context
func CtxUnlock(ctx context.Context) error {
	lock, err := GetLockFromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "get lock from ctx")
	}

	lock.Unlock()
	return nil
}

// CtxRLock lock the lock in gin context
func CtxRLock(ctx context.Context) error {
	lock, err := GetLockFromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "get lock from ctx")
	}

	lock.RLock()
	return nil
}

// CtxRUnlock unlock the lock in gin context
func CtxRUnlock(ctx context.Context) error {
	lock, err := GetLockFromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "get lock from ctx")
	}

	lock.RUnlock()
	return nil
}
