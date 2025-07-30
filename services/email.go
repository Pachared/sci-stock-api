package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

var (
	supportEmail    = "sci.next.stock@gmail.com"
	supportName     = "SCI Support"
	supportPassword string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("[INFO] .env not loaded, using system environment")
	}

	supportPassword = os.Getenv("SUPPORT_EMAIL_PASSWORD")
	supportPassword = strings.TrimSpace(supportPassword)
	supportPassword = strings.ReplaceAll(supportPassword, " ", "")

	if supportPassword == "" {
		fmt.Println("[ERROR] SUPPORT_EMAIL_PASSWORD is empty or not set!")
	}
}

func SendEmail(to, subject, htmlBody, plainText string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", supportEmail, supportName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer("smtp.gmail.com", 587, supportEmail, supportPassword)

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("[ERROR] SendEmail failed: %v\n", err)
		return err
	}
	return nil
}

func GenerateEmailBodyForOTP(otp string) (string, string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>OTP สำหรับเปลี่ยนรหัสผ่าน</title>
	</head>
	<body style="margin:0; padding:0; background-color:#F2F5F5;">
		<table align="center" border="0" cellpadding="0" cellspacing="0" width="100%%" style="background-color:#F2F5F5; padding:40px 0;">
			<tr>
				<td align="center">
					<table border="0" cellpadding="0" cellspacing="0" width="420" style="background:#ffffff; border-radius:12px; padding:40px; font-family:'Noto Sans Thai', sans-serif;">
						<tr>
							<td align="center" style="color:#000000; font-size:24px; font-weight:bold; padding-bottom:20px;">
								รหัส OTP สำหรับเปลี่ยนรหัสผ่าน
							</td>
						</tr>
						<tr>
							<td style="font-size:16px; color:#908E9B; text-align:center; padding-bottom:30px;">
								กรุณานำรหัส OTP ด้านล่างนี้ไปกรอกในระบบเพื่อรีเซ็ตรหัสผ่านของคุณ
							</td>
						</tr>
						<tr>
							<td align="center" style="background:#E1DFE9; color:#000000; font-size:32px; font-weight:bold; padding:20px 40px; border-radius:40px; letter-spacing:12px; margin-bottom:30px;">
								%s
							</td>
						</tr>
						<tr>
							<td align="center" style="font-size:14px; color:#888888; padding-top:20px; padding-bottom:30px;">
								รหัสนี้จะหมดอายุใน 10 นาที
							</td>
						</tr>
						<tr>
							<td align="center">
								<a href="#" style="background-color:#000000; color:#ffffff; padding:12px 30px; text-decoration:none; border-radius:20px; font-size:16px; font-weight:bold; display:inline-block;">เปลี่ยนรหัสผ่าน</a>
							</td>
						</tr>
						<tr>
							<td align="center" style="font-size:12px; color:#aaaaaa; padding-top:40px;">
								หากคุณไม่ได้ร้องขอ กรุณาเพิกเฉยต่ออีเมลนี้
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</body>
	</html>
	`, otp)

	plain := fmt.Sprintf("รหัส OTP สำหรับเปลี่ยนรหัสผ่านของคุณคือ: %s (หมดอายุใน 10 นาที)\nหากคุณไม่ได้ร้องขอ กรุณาเพิกเฉยต่ออีเมลนี้", otp)

	return html, plain
}

func GenerateEmailBodyForRegisterOTP(otp string) (string, string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="th">
	<head>
		<meta charset="UTF-8">
		<title>OTP ยืนยันอีเมล</title>
	</head>
	<body style="margin:0; padding:0; background-color:#F2F5F5;">
		<table align="center" border="0" cellpadding="0" cellspacing="0" width="100%%" style="background-color:#F2F5F5; padding:40px 0;">
			<tr>
				<td align="center">
					<table border="0" cellpadding="0" cellspacing="0" width="420" style="background:#ffffff; border-radius:12px; padding:40px; font-family:'Noto Sans Thai', sans-serif;">
						<tr>
							<td align="center" style="color:#000000; font-size:24px; font-weight:bold; padding-bottom:20px;">
								ยืนยันการสมัครสมาชิก
							</td>
						</tr>
						<tr>
							<td style="font-size:16px; color:#908E9B; text-align:center; padding-bottom:30px;">
								กรุณานำรหัส OTP ด้านล่างนี้ไปกรอกในระบบเพื่อยืนยันอีเมลของคุณ
							</td>
						</tr>
						<tr>
							<td align="center" style="background:#E1DFE9; color:#000000; font-size:32px; font-weight:bold; padding:20px 40px; border-radius:40px; letter-spacing:12px; margin-bottom:30px;">
								%s
							</td>
						</tr>
						<tr>
							<td align="center" style="font-size:14px; color:#888888; padding-top:20px; padding-bottom:30px;">
								รหัสนี้จะหมดอายุใน 10 นาที
							</td>
						</tr>
						<tr>
							<td align="center">
								<a href="#" style="background-color:#000000; color:#ffffff; padding:12px 30px; text-decoration:none; border-radius:20px; font-size:16px; font-weight:bold; display:inline-block;">เข้าสู่ระบบ</a>
							</td>
						</tr>
						<tr>
							<td align="center" style="font-size:12px; color:#aaaaaa; padding-top:40px;">
								ระบบคลังสินค้า SCI-Stock | %s
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</body>
	</html>
	`, otp, time.Now().Format("02/01/2006"))

	plain := fmt.Sprintf("OTP สำหรับยืนยันอีเมลของคุณคือ: %s (หมดอายุใน 10 นาที)", otp)

	return html, plain
}

