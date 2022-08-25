package interceptors

import (
	"context"

	"google.golang.org/grpc/tap"
)

//TapInHandleChain ...
type TapInHandleChain []tap.ServerInHandle

//TapInHandle ...
func (tc TapInHandleChain) TapInHandle(ctx context.Context, info *tap.Info) (context.Context, error) {
	var err error
	ctx2 := ctx
	for _, h := range tc {
		if ctx2, err = h(ctx2, info); err != nil {
			return ctx, err
		}
	}
	return ctx2, nil
}
