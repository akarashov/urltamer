package grpcserver

import (
	"context"

	th "github.com/akarashov/urltamer/internal/handler"
	pb "github.com/akarashov/urltamer/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServer implements the gRPC server for URL shortening service
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	h *th.Handler
}

// NewShortenerServer creates a new instance of ShortenerServer
func NewShortenerServer(h *th.Handler) *ShortenerServer {
	return &ShortenerServer{
		h: h,
	}
}

// AuthInterceptor is a gRPC interceptor(middleware) for authentication
func (s *ShortenerServer) AuthInterceptor(ctx context.Context, in interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userID int
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("authorization")
		if len(values) > 0 {
			tokenString := values[0]
			id := th.GetUserID(tokenString)
			if id > 0 {
				userID = id
			} else {
				return nil, status.Error(codes.Unauthenticated, "Invalid authorization token")
			}
		} else {
			// return nil, status.Error(codes.InvalidArgument, "Authorization token required")
		}
	}
	newCtx := context.WithValue(ctx, "uid", userID)
	return handler(newCtx, in)
}

// ShortenURL implements the ShortenerServiceServer interface
func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	var response pb.URLShortenResponse
	inputURL := in.GetUrl()
	if inputURL == "" {
		return nil, status.Error(codes.InvalidArgument, "URL cannot be empty")
	}
	uid := ctx.Value("uid").(int)
	tamer, _ := s.h.Service.CreateShortURL(ctx, inputURL, uid)
	response.SetResult(tamer.ShortURL)
	return &response, nil
}

// ExpandURL implements the ShortenerServiceServer interface
func (s *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	var response pb.URLExpandResponse
	tamer := in.GetId()
	if tamer == "" {
		return nil, status.Error(codes.InvalidArgument, "ID cannot be empty")
	}
	originalURL, err := s.h.Service.GetOriginalURL(ctx, tamer)
	if err != nil {
		return nil, status.Error(codes.NotFound, "URL not found")
	}
	response.SetResult(originalURL)
	return &response, nil
}

// ListUserURLs implements the ShortenerServiceServer interface
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	var response pb.UserURLsResponse
	var urlData pb.URLData
	uid := ctx.Value("uid").(int)
	userURLs, err := s.h.Service.GetUserURLs(ctx, uid)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve user URLs")
	}
	var pbUrls []*pb.URLData
	for _, url := range userURLs {
		urlData.SetShortUrl(url.ShortURL)
		urlData.SetOriginalUrl(url.OriginalURL)
		pbUrls = append(pbUrls, &urlData)
	}
	response.SetUrl(pbUrls)
	return &response, nil
}
