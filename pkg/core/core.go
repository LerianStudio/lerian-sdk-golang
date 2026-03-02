package core

// RawBody is a sentinel type that, when passed as the body argument to
// [Backend.Call], [Backend.CallWithHeaders], or [Backend.CallRaw], instructs
// the [BackendImpl] to send the contained bytes verbatim instead of
// JSON-marshaling them.
//
// This is used for non-JSON content types such as multipart/form-data.
// When using RawBody, the caller is also responsible for setting the
// correct Content-Type header via CallWithHeaders.
//
// Example:
//
//	headers := map[string]string{"Content-Type": writer.FormDataContentType()}
//	raw := core.RawBody{Data: buf.Bytes()}
//	err := backend.CallWithHeaders(ctx, "POST", "/templates", headers, raw, &result)
type RawBody struct {
	// Data is the pre-serialized request body bytes.
	Data []byte
}
