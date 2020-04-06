package http

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func (p *Protocol) do(ctx context.Context, req *http.Request) (binding.Message, error) {
	params := cecontext.RetriesFrom(ctx)

	switch params.Strategy {
	case cecontext.BackoffStrategyConstant, cecontext.BackoffStrategyLinear, cecontext.BackoffStrategyExponential:
		return p.doWithRetry(ctx, params)(req)
	case cecontext.BackoffStrategyNone:
		fallthrough
	default:
		return p.doOnce(req)
	}
}

func (p *Protocol) doOnce(req *http.Request) (binding.Message, protocol.Result) {
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, protocol.NewReceipt(false, "%w", err)
	}

	var result protocol.Result
	if resp.StatusCode/100 == 2 {
		result = protocol.ResultACK
	} else {
		result = protocol.ResultNACK
	}

	return NewMessage(resp.Header, resp.Body), NewResult(resp.StatusCode, "%w", result)
}

func (p *Protocol) doWithRetry(ctx context.Context, params *cecontext.RetryParams) func(*http.Request) (binding.Message, error) {
	return func(req *http.Request) (binding.Message, error) {
		params := params
		then := time.Now()
		retry := 0
		results := make([]protocol.Result, 0)

		for {
			msg, result := p.doOnce(req)

			// Fast track common case.
			if protocol.IsACK(result) {
				return msg, NewRetriesResult(result, retry, then, results)
			}

			// Try again?
			//
			// Make sure the error was something we should retry.

			//{
			//	var sErr *syscall.Errno
			//	if errors.As(result, &sErr) {
			//		// system level error, retry.
			//		goto DO_BACKOFF
			//	}
			//}

			{
				var uErr *url.Error
				if errors.As(result, &uErr) {
					// Only retry when the error is a timeout
					//if uErr.Timeout() || uErr.Temporary() {
					goto DO_BACKOFF
					//}
					//cecontext.LoggerFrom(ctx).Debugw("not timeout, will not try again", zap.Error(uErr))
					//return msg, NewRetriesResult(result, retry, then, results)
				}
			}

			{
				var httpResult *Result
				if errors.As(result, &httpResult) {
					// Potentially retry when:
					// - 404 Not Found
					// - 413 Payload Too Large with Retry-After (NOT SUPPORTED)
					// - 425 Too Early
					// - 429 Too Many Requests
					// - 503 Service Unavailable (with or without Retry-After) (IGNORE Retry-After)
					// - 504 Gateway Timeout

					sc := httpResult.StatusCode
					if sc == 404 || sc == 425 || sc == 429 || sc == 503 || sc == 504 {
						// retry!
						goto DO_BACKOFF
					} else {
						// Permanent error
						cecontext.LoggerFrom(ctx).Debugw("status code not retryable, will not try again",
							zap.Error(httpResult),
							zap.Int("statusCode", sc))
						return msg, NewRetriesResult(result, retry, then, results)
					}
				}
			}

		DO_BACKOFF:
			// Wait for the correct amount of backoff time.

			// total tries = retry + 1
			if err := params.Backoff(ctx, retry+1); err != nil {
				// do not try again.
				cecontext.LoggerFrom(ctx).Debugw("backoff error, will not try again", zap.Error(err))
				return msg, NewRetriesResult(result, retry, then, results)
			}

			retry++
			results = append(results, result)
		}
	}
}
