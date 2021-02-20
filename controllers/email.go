package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"rewan/models"
	"strconv"
)

//Function used to send a simple email to the configured receiver after verifying the captcha
//(Here the sender also is the receiver)
func SendMail() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Print("email.go > SendMail")
		//Data to be sent to the ReCaptcha service
		data := url.Values{
			//Secret key, fetched from the configuration
			"secret": {models.Conf.CaptchaKey},
			//Captcha data completed by the user, fetched from the POST data
			"response": {c.PostForm("recaptcha")},
		}

		//Send the data to the ReCaptcha service and get the response
		resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", data)

		//If there is an error here, we set a flash error message
		//We also exit the function just after calling the route,
		//to avoid getting a decoding error later on and processing useless actions
		if err != nil {
			log.Print(MsgStr("mail.error.connect", "en", nil))
			Flashvar("error", "mail.error.connect", nil, c)
			c.HTML(http.StatusOK, "contact.html", Render(c, gin.H{"title": "contact"}))
			return
		}

		var res map[string]interface{}

		//Decoding the response we got earlier
		_ = json.NewDecoder(resp.Body).Decode(&res)

		//If the response indicates a success, we send the mail
		if res["success"] == true {
			//We set all the SMTP data
			from := models.Conf.SmtpSender
			password := models.Conf.SmtpPassword
			to := []string{
				models.Conf.SmtpSender,
			}
			host := models.Conf.SmtpHost
			port := models.Conf.SmtpPort
			//Since I'm the only one to see this mail,
			//the subject and content do not need a translation,
			//but it would otherwise be a great idea to call controllers.MsgStr here
			subject := "Message depuis le site !!"
			//We here set the body of the message, using the POST data content
			content := "Envoyé par : " + c.PostForm("mail") + "\nMessage : " + c.PostForm("content")
			//This method simplifies the generation of the mail and its headers
			message := getMailContent(models.Conf.SmtpSender, models.Conf.SmtpSender, subject, content)

			//We use the already defined variables to create an Auth object to use
			auth := smtp.PlainAuth("", from, password, host)
			//We then use the Auth object and the other variables to
			errMail := smtp.SendMail(host+":"+strconv.Itoa(port), auth, from, to, []byte(message))
			if errMail != nil {
				//If we get this error, it means the email was not sent
				log.Print(MsgStr("mail.error.notsent", "en", nil))
				Flashvar("error", "mail.error.notsent", nil, c)
			} else {
				//Else, we notify the user that the email was sent
				log.Print(MsgStr("mail.success.sent", "en", nil))
				Flashvar("success", "mail.success.sent", nil, c)
			}
		} else {
			//If we get this error, it means that the captcha was not completed successfully
			log.Print(MsgStr("mail.error.captcha", "en", nil))
			Flashvar("error", "mail.error.captcha", nil, c)
		}
		//Then we get the contact page
		c.HTML(http.StatusOK, "contact.html", Render(c, gin.H{"title": "contact"}))
	}
}

//This function simplifies the email generation
//by translating standard arguments (sender, receiver, subject, content) to an email
func getMailContent(sender string, receiver string, subject string, content string) string {
	log.Printf("email.go > getMailContent [ %v / %v / %v / %v ]", sender, receiver, subject, content)
	msg := "To: <" + models.Conf.SmtpSender + ">\nFrom: <" + models.Conf.SmtpSender + ">\nSubject: " + subject + "\n" + content
	return msg
}
