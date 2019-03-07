package context

import (
	"context"
	"net/url"
)

// Opaque key type used to store target
type targetKeyType struct{}

var targetKey = targetKeyType{}

func ContextWithTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, targetKey, target)
}

func TargetFromContext(ctx context.Context) *url.URL {
	c := ctx.Value(targetKey)
	if c != nil {
		if target, err := url.Parse(c.(string)); err == nil {
			return target
		}
	}
	return nil
}

//
//// Opaque key type used to store port
//type portKeyType struct{}
//
//var portKey = portKeyType{}
//
//func ContextWithPort(ctx context.Context, port int) context.Context {
//	return context.WithValue(ctx, portKey, port)
//}
//
//func PortFromContext(ctx context.Context) int {
//	v := ctx.Value(portKey)
//	if v != nil {
//		if port, ok := v.(int); ok {
//			return port
//		}
//	}
//	return 0
//}
