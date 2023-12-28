package jwtcore

import "github.com/golang-jwt/jwt/v5"

type Claims[T jwt.Claims] interface {
	jwt.Claims

	SetIssuer(issuer string)
	SetSubject(subject string)
	SetAudience(audience jwt.ClaimStrings)
	SetExpiresAt(expiresAt *jwt.NumericDate)
	SetNotBefore(notBefore *jwt.NumericDate)
	SetIssuedAt(issuedAt *jwt.NumericDate)
	SetID(id string)
	*T
}
