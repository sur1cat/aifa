package handler

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/sur1cat/aifa/auth-service/internal/domain"
	"github.com/sur1cat/aifa/auth-service/internal/otp"

	"github.com/gin-gonic/gin"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type OTPHandler struct {
	auth *AuthHandler
	otp  *otp.Store
	debug bool
}

func NewOTPHandler(auth *AuthHandler, otpStore *otp.Store, debug bool) *OTPHandler {
	return &OTPHandler{auth: auth, otp: otpStore, debug: debug}
}

type sendOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *OTPHandler) SendOTP(c *gin.Context) {
	var req sendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	if !phoneRegex.MatchString(req.Phone) {
		respondError(c, http.StatusBadRequest, codeValidation, "Invalid phone number format, expected +<country><number>")
		return
	}

	code, expiresAt, err := h.otp.Generate(c.Request.Context(), req.Phone)
	if err != nil {
		slog.Error("otp generate", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to send OTP")
		return
	}

	if h.debug {
		slog.Info("OTP generated (debug mode only)", "phone", req.Phone, "code", code)
	}

	respondOK(c, gin.H{
		"message":    "OTP sent successfully",
		"expires_at": expiresAt,
	})
}

type verifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	if !phoneRegex.MatchString(req.Phone) {
		respondError(c, http.StatusBadRequest, codeValidation, "Invalid phone number format")
		return
	}

	valid, err := h.otp.Verify(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		slog.Error("otp verify", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to verify OTP")
		return
	}

	if !valid {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Invalid or expired OTP code")
		return
	}

	h.auth.signIn(c, domain.ProviderPhone, providerProfile{
		Sub: req.Phone,
	})
}
