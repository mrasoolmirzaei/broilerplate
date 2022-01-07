package mail

import (
	"bytes"
	"fmt"
	conf "github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/models"
	"github.com/muety/broilerplate/routes"
	"github.com/muety/broilerplate/services"
	"github.com/muety/broilerplate/utils"
	"github.com/muety/broilerplate/views/mail"
)

const (
	tplNamePasswordReset = "reset_password"
	subjectPasswordReset = "Broilerplate - Password Reset"
)

type SendingService interface {
	Send(*models.Mail) error
}

type MailService struct {
	config         *conf.Config
	sendingService SendingService
	templates      utils.TemplateMap
}

func NewMailService() services.IMailService {
	config := conf.Get()

	var sendingService SendingService
	sendingService = &NoopSendingService{}

	if config.Mail.Enabled {
		if config.Mail.Provider == conf.MailProviderMailWhale {
			sendingService = NewMailWhaleSendingService(config.Mail.MailWhale)
		} else if config.Mail.Provider == conf.MailProviderSmtp {
			sendingService = NewSMTPSendingService(config.Mail.Smtp)
		}
	}

	// Use local file system when in 'dev' environment, go embed file system otherwise
	templateFs := conf.ChooseFS("views/mail", mail.TemplateFiles)
	templates, err := utils.LoadTemplates(templateFs, routes.DefaultTemplateFuncs())
	if err != nil {
		panic(err)
	}

	return &MailService{sendingService: sendingService, config: config, templates: templates}
}

func (m *MailService) SendPasswordReset(recipient *models.User, resetLink string) error {
	tpl, err := m.getPasswordResetTemplate(PasswordResetTplData{ResetLink: resetLink})
	if err != nil {
		return err
	}
	mail := &models.Mail{
		From:    models.MailAddress(m.config.Mail.Sender),
		To:      models.MailAddresses([]models.MailAddress{models.MailAddress(recipient.Email)}),
		Subject: subjectPasswordReset,
	}
	mail.WithHTML(tpl.String())
	return m.sendingService.Send(mail)
}

func (m *MailService) getPasswordResetTemplate(data PasswordResetTplData) (*bytes.Buffer, error) {
	var rendered bytes.Buffer
	if err := m.templates[m.fmtName(tplNamePasswordReset)].Execute(&rendered, data); err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (m *MailService) fmtName(name string) string {
	return fmt.Sprintf("%s.tpl.html", name)
}
