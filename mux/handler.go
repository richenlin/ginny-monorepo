package mux

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

const (
	statusOKPrefix         = 2
	statusBadRequestPrefix = 4
	fallback               = `{"code": 13,"message":"failed to marshal error message"}`
)

// codesErrors some errors string for grpc codes
var codesErrors = map[codes.Code]string{
	codes.OK:                 "ok",
	codes.Canceled:           "canceled",
	codes.Unknown:            "unknown",
	codes.InvalidArgument:    "invalid_argument",
	codes.DeadlineExceeded:   "deadline_exceeded",
	codes.NotFound:           "not_found",
	codes.AlreadyExists:      "already_exists",
	codes.PermissionDenied:   "permission_denied",
	codes.ResourceExhausted:  "resource_exhausted",
	codes.FailedPrecondition: "failed_precondition",
	codes.Aborted:            "aborted",
	codes.OutOfRange:         "out_of_range",
	codes.Unimplemented:      "unimplemented",
	codes.Internal:           "internal",
	codes.Unavailable:        "unavailable",
	codes.DataLoss:           "data_loss",
	codes.Unauthenticated:    "unauthenticated",
}

// httpStatusCode the 2xxx is 200, the 4xxx is 400, the 5xxx is 500
func httpStatusCode(code codes.Code) (httpStatusCode int) {
	// http status codes can be error codes
	if code >= 200 && code < 599 {
		return int(code)
	}
	for code >= 10 {
		code /= 10
	}
	switch code {
	case statusOKPrefix:
		httpStatusCode = http.StatusOK
	case statusBadRequestPrefix:
		httpStatusCode = http.StatusBadRequest
	default:
		httpStatusCode = http.StatusInternalServerError
	}
	return
}

// CodeToError
func CodeToError(c codes.Code) string {
	errStr, ok := codesErrors[c]
	if ok {
		return errStr
	}
	return strconv.FormatInt(int64(c), 10)
}

func codeToStatus(code codes.Code) int {
	st := int(code)
	if st > 100 {
		st = httpStatusCode(code)
	} else {
		st = runtime.HTTPStatusFromCode(code)
	}
	return st
}

func defaultErrorHandler(ctx context.Context,
	mux *runtime.ServeMux, marshaler runtime.Marshaler,
	w http.ResponseWriter, r *http.Request, err error,
) {
	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("Failed to extract ServerMetadata from context")
	}

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.

	if te := r.Header.Get("TE"); strings.Contains(strings.ToLower(te), "trailers") {
		handleForwardResponseTrailerHeader(w, md)
		w.Header().Set("Transfer-Encoding", "chunked")
		handleForwardResponseTrailer(w, md)
	}

	WriteHTTPErrorResponse(w, r, err)
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := runtime.MetadataTrailerPrefix + k
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

// WriteHTTPErrorResponse  set HTTP status code and write error description to the body.
func WriteHTTPErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	preTags := tags.Extract(r.Context())
	preTags.Set("code", strconv.Itoa(int(s.Code())))
	preTags.Set("message", s.Message())
	w.WriteHeader(codeToStatus(s.Code()))

	_, e := defaultOptions.bodyWriter(w, r, s)
	// newErrorBytes(requestId, apiModel, s, optsWare.errorMarshaler, optsWare.newErrorBody)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}
}

// handlerWithMiddleWares handler with middle wares.
func handlerWithMiddleWares(h http.Handler, middleWares ...func(http.Handler) http.Handler) http.Handler {
	lenMiddleWare := len(middleWares)
	for i := lenMiddleWare - 1; i >= 0; i-- {
		middleWare := middleWares[i]
		h = middleWare(h)
	}
	return h
}
