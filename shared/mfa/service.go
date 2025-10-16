package mfa

type Service struct {
	TOTP *TOTPService
	OTP  *OTPService
}

func NewService(totp *TOTPService, otp *OTPService) *Service {
	return &Service{
		TOTP: totp,
		OTP:  otp,
	}
}
