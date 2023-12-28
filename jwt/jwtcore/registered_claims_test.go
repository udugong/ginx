package jwtcore

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredClaims_GetAudience(t *testing.T) {
	tests := []struct {
		name    string
		want    jwt.ClaimStrings
		wantErr error
	}{
		{
			name:    "normal",
			want:    jwt.ClaimStrings{"bar"},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetAudience()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_GetExpirationTime(t *testing.T) {
	tests := []struct {
		name    string
		want    *jwt.NumericDate
		wantErr error
	}{
		{
			name:    "normal",
			want:    jwt.NewNumericDate(nowTime),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetExpirationTime()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_GetIssuedAt(t *testing.T) {
	tests := []struct {
		name    string
		want    *jwt.NumericDate
		wantErr error
	}{
		{
			name:    "normal",
			want:    jwt.NewNumericDate(nowTime),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetIssuedAt()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_GetIssuer(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr error
	}{
		{
			name:    "normal",
			want:    "foo",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetIssuer()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_GetNotBefore(t *testing.T) {
	tests := []struct {
		name    string
		want    *jwt.NumericDate
		wantErr error
	}{
		{
			name:    "normal",
			want:    jwt.NewNumericDate(nowTime),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetNotBefore()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_GetSubject(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr error
	}{
		{
			name:    "normal",
			want:    "foo",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			got, err := c.GetSubject()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisteredClaims_SetAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience jwt.ClaimStrings
		want     jwt.ClaimStrings
	}{
		{
			name:     "normal",
			audience: jwt.ClaimStrings{"123"},
			want:     jwt.ClaimStrings{"123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetAudience(tt.audience)
			assert.Equal(t, tt.want, c.Audience)
		})
	}
}

func TestRegisteredClaims_SetExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *jwt.NumericDate
		want      *jwt.NumericDate
	}{
		{
			name:      "normal",
			expiresAt: jwt.NewNumericDate(nowTime.Add(time.Hour)),
			want:      jwt.NewNumericDate(nowTime.Add(time.Hour)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetExpiresAt(tt.expiresAt)
			assert.Equal(t, tt.want, c.ExpiresAt)
		})
	}
}

func TestRegisteredClaims_SetID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "normal",
			id:   "123",
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetID(tt.id)
			assert.Equal(t, tt.want, c.ID)
		})
	}
}

func TestRegisteredClaims_SetIssuedAt(t *testing.T) {
	tests := []struct {
		name     string
		issuedAt *jwt.NumericDate
		want     *jwt.NumericDate
	}{
		{
			name:     "normal",
			issuedAt: jwt.NewNumericDate(nowTime.Add(time.Minute)),
			want:     jwt.NewNumericDate(nowTime.Add(time.Minute)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetIssuedAt(tt.issuedAt)
			assert.Equal(t, tt.want, c.IssuedAt)
		})
	}
}

func TestRegisteredClaims_SetIssuer(t *testing.T) {
	tests := []struct {
		name   string
		issuer string
		want   string
	}{
		{
			name:   "normal",
			issuer: "udugong",
			want:   "udugong",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetIssuer(tt.issuer)
			assert.Equal(t, tt.want, c.Issuer)
		})
	}
}

func TestRegisteredClaims_SetNotBefore(t *testing.T) {
	tests := []struct {
		name      string
		notBefore *jwt.NumericDate
		want      *jwt.NumericDate
	}{
		{
			name:      "normal",
			notBefore: jwt.NewNumericDate(nowTime.Add(time.Minute)),
			want:      jwt.NewNumericDate(nowTime.Add(time.Minute)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetNotBefore(tt.notBefore)
			assert.Equal(t, tt.want, c.NotBefore)
		})
	}
}

func TestRegisteredClaims_SetSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    string
	}{
		{
			name:    "normal",
			subject: "udugong",
			want:    "udugong",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := defClaims
			c.SetSubject(tt.subject)
			assert.Equal(t, tt.want, c.Subject)
		})
	}
}

var defClaims = RegisteredClaims{
	Issuer:    "foo",
	Subject:   "foo",
	Audience:  jwt.ClaimStrings{"bar"},
	ExpiresAt: jwt.NewNumericDate(nowTime),
	NotBefore: jwt.NewNumericDate(nowTime),
	IssuedAt:  jwt.NewNumericDate(nowTime),
	ID:        "bar",
}
