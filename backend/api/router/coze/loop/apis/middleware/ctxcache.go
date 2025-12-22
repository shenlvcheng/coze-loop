// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/coze-dev/coze-loop/backend/modules/evaluation/consts"
	"github.com/coze-dev/coze-loop/backend/pkg/ctxcache"
)

func CtxCacheMW() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ctx = ctxcache.Init(ctx)

		if val := c.GetHeader(consts.ZhiyuAuthorizationHeader); len(val) > 0 {
			ctxcache.Store(ctx, consts.ZhiyuAuthorizationCtxKey, string(val))
		}
		if val := c.GetHeader(consts.ZhiyuAuthTokenHeader); len(val) > 0 {
			ctxcache.Store(ctx, consts.ZhiyuAuthTokenCtxKey, string(val))
		}
		if val := c.GetHeader(consts.ZhiyuSceneTypeHeader); len(val) > 0 {
			ctxcache.Store(ctx, consts.ZhiyuSceneTypeCtxKey, string(val))
		}

		c.Next(ctx)
	}
}
