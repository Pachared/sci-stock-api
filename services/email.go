package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func GenerateEmailBodyForOTP(otp string) (string, string) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="th">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>รหัส OTP ของคุณ</title>
</head>
<body style="font-family: 'Helvetica', Arial, sans-serif; background-color: #f4f6f8; margin: 0; padding: 0;">
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background-color: #f4f6f8; padding: 40px 0;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%%" max-width="600px" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 10px; padding: 40px; box-shadow: 0 4px 20px rgba(0,0,0,0.1);">
                    <tr>
                        <td align="center">
                            <h1 style="color: #333333; font-size: 24px; margin-bottom: 20px;">รหัส OTP ของคุณ</h1>
                            <div style="display: inline-block; background-color: #2a9d8f; color: #ffffff; font-size: 32px; font-weight: bold; padding: 15px 30px; border-radius: 8px; letter-spacing: 4px; margin-bottom: 20px;">
                                %s
                            </div>
                            <p style="color: #555555; font-size: 16px; line-height: 1.5; margin-bottom: 30px;">รหัสนี้จะหมดอายุใน <strong>10 นาที</strong> กรุณาอย่าแชร์รหัสนี้กับผู้อื่น</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, otp)

	plain := fmt.Sprintf("รหัส OTP ของคุณ: %s (หมดอายุใน 10 นาที)", otp)
	return html, plain
}

func GenerateEmailBodyForRegisterOTP(otp string) (string, string) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="th">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>ยืนยันการสมัครสมาชิก</title>
</head>
<body style="font-family: 'Helvetica', Arial, sans-serif; background-color: #f4f6f8; margin: 0; padding: 0;">
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background-color: #f4f6f8; padding: 40px 0;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%%" max-width="600px" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 10px; padding: 40px; box-shadow: 0 4px 20px rgba(0,0,0,0.1);">
                    <tr>
                        <td align="center">
                            <h1 style="color: #333333; font-size: 24px; margin-bottom: 20px;">ยืนยันการสมัครสมาชิก</h1>
                            <div style="display: inline-block; background-color: #e76f51; color: #ffffff; font-size: 32px; font-weight: bold; padding: 15px 30px; border-radius: 8px; letter-spacing: 4px; margin-bottom: 20px;">
                                %s
                            </div>
                            <p style="color: #555555; font-size: 16px; line-height: 1.5; margin-bottom: 30px;">
                                รหัสนี้จะหมดอายุใน <strong>10 นาที</strong> กรุณาอย่าแชร์รหัสนี้กับผู้อื่น
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, otp)

	plain := fmt.Sprintf("OTP สำหรับสมัครสมาชิกคือ: %s (หมดอายุใน 10 นาที)", otp)
	return html, plain
}

func GoogleSendMail(to, subject, htmlBody, plainBody string) error {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("ไม่พบไฟล์ credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return fmt.Errorf("โหลด config ไม่สำเร็จ: %v", err)
	}

	tok, err := tokenFromFile("token.json")
	if err != nil {
		return fmt.Errorf("ไม่พบ token.json: %v", err)
	}

	client := config.Client(context.Background(), tok)
	srv, err := gmail.New(client)
	if err != nil {
		return fmt.Errorf("สร้าง Gmail service ไม่สำเร็จ: %v", err)
	}

	from := "sci.next.stock@gmail.com"

	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	raw := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from,
		to,
		encodedSubject,
		htmlBody,
	)

	message := &gmail.Message{
		Raw: base64.StdEncoding.EncodeToString([]byte(raw)),
	}

	_, err = srv.Users.Messages.Send("me", message).Do()
	if err != nil {
		return fmt.Errorf("ส่งอีเมลล้มเหลว: %v", err)
	}

	return nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	var code string
	fmt.Printf("Enter the code from that page here: ")
	fmt.Scan(&code)

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

func SaveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("ไม่สามารถบันทึก token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}