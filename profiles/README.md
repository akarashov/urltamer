File: main
Type: inuse_space
Time: 2026-01-22 22:17:59 MSK
Showing nodes accounting for 203.60MB, 585.79% of 34.76MB total
Dropped 56 nodes (cum <= 0.17MB)
      flat  flat%   sum%        cum   cum%
  168.35MB 484.39% 484.39%   203.60MB 585.79%  compress/flate.NewWriter (inline)
   35.24MB 101.40% 585.79%    35.24MB 101.40%  compress/flate.(*compressor).initDeflate (inline)
         0     0% 585.79%    35.24MB 101.40%  compress/flate.(*compressor).init
         0     0% 585.79%    -4.41MB 12.68%  compress/gzip.(*Writer).Close
         0     0% 585.79%   203.60MB 585.79%  compress/gzip.(*Writer).Write
         0     0% 585.79%   141.28MB 406.49%  fmt.Fprintln
         0     0% 585.79%    28.23MB 81.23%  github.com/akarashov/urltamer/internal/handler.(*Handler).DeleteUserURLsEndpoint
         0     0% 585.79%    36.57MB 105.22%  github.com/akarashov/urltamer/internal/handler.(*Handler).PingEndpoint
         0     0% 585.79%    24.93MB 71.72%  github.com/akarashov/urltamer/internal/handler.(*Handler).RequestEndpoint
         0     0% 585.79%    40.80MB 117.39%  github.com/akarashov/urltamer/internal/handler.(*Handler).RequestJSONEndpoint
         0     0% 585.79%    34.45MB 99.11%  github.com/akarashov/urltamer/internal/handler.(*Handler).RequestJSONEndpointBatch
         0     0% 585.79%    42.53MB 122.37%  github.com/akarashov/urltamer/internal/handler.(*Handler).ResponseEndpoint
         0     0% 585.79%    -4.41MB 12.68%  github.com/akarashov/urltamer/internal/handler.(*compressWriter).Close
         0     0% 585.79%      208MB 598.47%  github.com/akarashov/urltamer/internal/handler.(*compressWriter).Write
         0     0% 585.79%   232.34MB 668.48%  github.com/akarashov/urltamer/internal/handler.(*loggingResponseWriter).Write
         0     0% 585.79%   233.84MB 672.79%  github.com/akarashov/urltamer/internal/handler.CookieMiddleware.func1
         0     0% 585.79%   233.84MB 672.79%  github.com/akarashov/urltamer/internal/handler.GzipMiddleware.func1
         0     0% 585.79%   204.10MB 587.23%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 585.79%   203.10MB 584.35%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 585.79%    -3.14MB  9.05%  main.main.AuditMiddleware.func5
         0     0% 585.79%    28.07MB 80.76%  main.main.AuditMiddleware.func7
         0     0% 585.79%    40.80MB 117.39%  main.main.AuditMiddleware.func8
         0     0% 585.79%    42.53MB 122.37%  main.main.AuditMiddleware.func9
         0     0% 585.79%    -3.14MB  9.05%  main.main.CookieMiddleware.func2
         0     0% 585.79%    -8.46MB 24.35%  main.main.CookieMiddleware.func20
         0     0% 585.79%    -4.41MB 12.68%  main.main.GzipMiddleware.func14
         0     0% 585.79%    -8.46MB 24.35%  main.main.GzipMiddleware.func21
         0     0% 585.79%    -3.14MB  9.05%  main.main.GzipMiddleware.func3
         0     0% 585.79%    -4.03MB 11.58%  main.main.GzipMiddleware.func6
         0     0% 585.79%    -9.70MB 27.90%  main.main.GzipMiddleware.func9
         0     0% 585.79%    -9.70MB 27.90%  main.main.LoggingMiddlewareRequest.func10
         0     0% 585.79%    -8.46MB 24.35%  main.main.LoggingMiddlewareRequest.func22
         0     0% 585.79%    -3.14MB  9.05%  main.main.LoggingMiddlewareRequest.func4
         0     0% 585.79%    -4.03MB 11.58%  main.main.LoggingMiddlewareRequest.func7
         0     0% 585.79%    -4.41MB 12.68%  main.main.LoggingMiddlewareResponse.func15
         0     0% 585.79%   233.34MB 671.35%  main.main.func2.1
         0     0% 585.79%   233.34MB 671.35%  main.main.func2.1.LoggingMiddlewareRequest.1 (inline)
         0     0% 585.79%   233.34MB 671.35%  main.main.func3.1
         0     0% 585.79%   233.34MB 671.35%  main.main.func3.1.LoggingMiddlewareResponse.1 (inline)
         0     0% 585.79%   233.84MB 672.79%  main.main.main.func1.func5.1
         0     0% 585.79%   233.84MB 672.79%  main.main.main.func1.func6.1
         0     0% 585.79%   204.60MB 588.66%  net/http.(*conn).serve
         0     0% 585.79%   141.78MB 407.93%  net/http.Error
         0     0% 585.79%   204.10MB 587.23%  net/http.HandlerFunc.ServeHTTP
         0     0% 585.79%   204.10MB 587.23%  net/http.serverHandler.ServeHTTP
