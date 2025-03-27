package middlewares

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLockableMw(t *testing.T) {
	gctx := &gin.Context{}
	mw := LockableMw()
	mw(gctx)

	lock, err := GetLockFromCtx(gctx)
	require.NoError(t, err)
	require.NotNil(t, lock)
}

func TestGetLockFromCtx(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		lock, err := GetLockFromCtx(nil)
		require.Error(t, err)
		require.Nil(t, lock)
	})

	t.Run("non-gin context", func(t *testing.T) {
		ctx := context.Background()
		lock, err := GetLockFromCtx(ctx)
		require.Error(t, err)
		require.Nil(t, lock)
	})

	t.Run("gin context without lock", func(t *testing.T) {
		gctx := &gin.Context{}
		lock, err := GetLockFromCtx(gctx)
		require.Error(t, err)
		require.Nil(t, lock)
	})

	t.Run("valid gin context with lock", func(t *testing.T) {
		gctx := &gin.Context{}
		mw := LockableMw()
		mw(gctx)

		lock, err := GetLockFromCtx(gctx)
		require.NoError(t, err)
		require.NotNil(t, lock)
	})
}

func TestLockOperations(t *testing.T) {
	gctx := &gin.Context{}
	mw := LockableMw()
	mw(gctx)

	t.Run("write lock operations", func(t *testing.T) {
		require.NoError(t, CtxLock(gctx))
		require.NoError(t, CtxUnlock(gctx))
	})

	t.Run("read lock operations", func(t *testing.T) {
		require.NoError(t, CtxRLock(gctx))
		require.NoError(t, CtxRUnlock(gctx))
	})

	t.Run("lock operations with nil context", func(t *testing.T) {
		require.Error(t, CtxLock(nil))
		require.Error(t, CtxUnlock(nil))
		require.Error(t, CtxRLock(nil))
		require.Error(t, CtxRUnlock(nil))
	})

	t.Run("invalid lock type in context", func(t *testing.T) {
		gctx := &gin.Context{}
		// Set invalid lock type
		gctx.Set(string(CtxKeyLock), "not a mutex")

		// Test all lock operations with invalid lock type
		lock, err := GetLockFromCtx(gctx)
		require.Error(t, err)
		require.Nil(t, lock)
		require.Contains(t, err.Error(), "is not a *sync.RWMutex")

		require.Error(t, CtxLock(gctx))
		require.Error(t, CtxUnlock(gctx))
		require.Error(t, CtxRLock(gctx))
		require.Error(t, CtxRUnlock(gctx))
	})
}
