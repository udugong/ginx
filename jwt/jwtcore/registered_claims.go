package jwtcore

import "github.com/golang-jwt/jwt/v5"

type RegisteredClaims struct {
	// the `iss` (Issuer) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1
	Issuer string `json:"iss,omitempty"`

	// the `sub` (Subject) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2
	Subject string `json:"sub,omitempty"`

	// the `aud` (Audience) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3
	Audience jwt.ClaimStrings `json:"aud,omitempty"`

	// the `exp` (Expiration Time) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4
	ExpiresAt *jwt.NumericDate `json:"exp,omitempty"`

	// the `nbf` (Not Before) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5
	NotBefore *jwt.NumericDate `json:"nbf,omitempty"`

	// the `iat` (Issued At) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6
	IssuedAt *jwt.NumericDate `json:"iat,omitempty"`

	// the `jti` (JWT ID) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7
	ID string `json:"jti,omitempty"`
}

// GetExpirationTime implements the Claims interface.
func (c RegisteredClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

// GetNotBefore implements the Claims interface.
func (c RegisteredClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.NotBefore, nil
}

// GetIssuedAt implements the Claims interface.
func (c RegisteredClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

// GetAudience implements the Claims interface.
func (c RegisteredClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.Audience, nil
}

// GetIssuer implements the Claims interface.
func (c RegisteredClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

// GetSubject implements the Claims interface.
func (c RegisteredClaims) GetSubject() (string, error) {
	return c.Subject, nil
}

func (c *RegisteredClaims) SetIssuer(issuer string) {
	c.Issuer = issuer
}

func (c *RegisteredClaims) SetSubject(subject string) {
	c.Subject = subject
}

func (c *RegisteredClaims) SetAudience(audience jwt.ClaimStrings) {
	c.Audience = audience
}

func (c *RegisteredClaims) SetExpiresAt(expiresAt *jwt.NumericDate) {
	c.ExpiresAt = expiresAt
}

func (c *RegisteredClaims) SetNotBefore(notBefore *jwt.NumericDate) {
	c.NotBefore = notBefore
}

func (c *RegisteredClaims) SetIssuedAt(issuedAt *jwt.NumericDate) {
	c.IssuedAt = issuedAt
}

func (c *RegisteredClaims) SetID(id string) {
	c.ID = id
}
