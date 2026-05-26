// Copyright 2026 The go-token-exchange Authors
// SPDX-License-Identifier: Apache-2.0

package tokenexchange

// GrantTypeTokenExchange is the OAuth 2.0 grant-type URI for token
// exchange. It is sent in the request's grant_type parameter
// (RFC 8693 §2.1) and is the only valid value for that field.
const GrantTypeTokenExchange = "urn:ietf:params:oauth:grant-type:token-exchange"

// Token-type URIs defined by RFC 8693 §3.
//
// These URIs identify the type of a security token carried in
// subject_token, actor_token, or access_token. The set below is the
// snapshot registered at the time RFC 8693 was published; the IANA
// "OAuth URI" registry grows additively, so consumers may encounter
// URIs not listed here. Such URIs are not errors — see
// [UnknownTokenType] and [RegisterTokenType].
const (
	// TokenTypeAccessToken indicates an OAuth 2.0 access token (RFC 6749).
	TokenTypeAccessToken = "urn:ietf:params:oauth:token-type:access_token"

	// TokenTypeRefreshToken indicates an OAuth 2.0 refresh token (RFC 6749).
	TokenTypeRefreshToken = "urn:ietf:params:oauth:token-type:refresh_token"

	// TokenTypeIDToken indicates an OpenID Connect ID Token.
	TokenTypeIDToken = "urn:ietf:params:oauth:token-type:id_token"

	// TokenTypeSAML1 indicates a base64url-encoded SAML 1.1 assertion.
	TokenTypeSAML1 = "urn:ietf:params:oauth:token-type:saml1"

	// TokenTypeSAML2 indicates a base64url-encoded SAML 2.0 assertion.
	TokenTypeSAML2 = "urn:ietf:params:oauth:token-type:saml2"

	// TokenTypeJWT indicates a JSON Web Token (RFC 7519).
	TokenTypeJWT = "urn:ietf:params:oauth:token-type:jwt"
)

// BuiltinTokenTypes returns the token-type URIs RFC 8693 §3 registered
// at the time of publication. The returned slice is a fresh copy and
// safe to mutate; the order matches the order the URIs appear in
// RFC 8693 §3.
func BuiltinTokenTypes() []string {
	return []string{
		TokenTypeAccessToken,
		TokenTypeRefreshToken,
		TokenTypeIDToken,
		TokenTypeSAML1,
		TokenTypeSAML2,
		TokenTypeJWT,
	}
}

// UnknownTokenType wraps a token-type URI the library does not
// recognize. It exists so that a request or response using a URI
// registered after this library was built round-trips verbatim
// rather than failing to parse.
//
// The wire form is the URI itself, unchanged. Library code that
// receives an UnknownTokenType has three options: treat it as
// opaque and pass through, call [RegisterTokenType] so subsequent
// parses recognize it, or reject it at the application layer.
type UnknownTokenType struct {
	// URI is the unrecognized token-type URI exactly as it appeared
	// on the wire.
	URI string
}

// String returns the wrapped URI, satisfying [fmt.Stringer]. It is
// equivalent to reading the URI field directly.
func (u UnknownTokenType) String() string {
	return u.URI
}
