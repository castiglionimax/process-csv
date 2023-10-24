package repository

import (
	"context"
	"fmt"
	"github.com/castiglionimax/process-csv/internal/domain"
	"github.com/harranali/mailing"
	"log"
	"strings"
	"time"
)

const (
	smtpServer = "localhost"
	smtpPort   = "25"
	from       = "sender@domain-poc.com"

	getSummary = "SELECT email,amount,period,credit,credit_qty,debit,debit_qty, summaries.last_updated FROM summaries  LEFT JOIN accounts ON summaries.account_id = accounts.id"
)

type (
	dao struct {
		Credit, Debit, Amount float32
		CreditQty, DebitQty   int
		Period, Email         string
		LastUpdated           time.Time
	}

	HtmlElement struct {
		name, text string
		elements   []HtmlElement
	}
)

const (
	indentSize = 2
)

func (r Repository) sendNotification(ctx context.Context, accountID domain.AccountID, msg []dao) error {

	r.mailer.SetFrom(mailing.EmailAddress{
		Name:    "sender name",
		Address: from,
	})

	r.mailer.SetTo([]mailing.EmailAddress{
		{Address: msg[0].Email},
	})

	// Set the subject
	r.mailer.SetSubject("This is the subject")
	_ = htmlBuilder(accountID, msg)
	r.mailer.SetHTMLBody("<h1>This is the email body</h1>")

	return r.mailer.Send()
}

func (r Repository) SendEmail(ctx context.Context, accountID domain.AccountID, start, end time.Time) error {
	var arrayPeriods []any
	for auxStar := start; auxStar.Before(end); auxStar = auxStar.AddDate(0, 1, 0) {
		year, month, _ := auxStar.Date()
		arrayPeriods = append(arrayPeriods, fmt.Sprintf("%d %s", year, month))
	}
	var query = getSummary + " WHERE period IN (?"

	placeholders := make([]any, len(arrayPeriods))
	for i, period := range arrayPeriods {
		if i > 0 {
			query += ", ?"
		}
		placeholders[i] = period
	}
	query += fmt.Sprintf(") AND id = ?;")
	arrayPeriods = append(arrayPeriods, string(accountID))

	rows, err := r.mysql.QueryContext(ctx, query, arrayPeriods...)
	if err != nil {
		return err
	}
	defer rows.Close()

	resp := make([]dao, 0)

	for rows.Next() {
		result := new(dao)
		if err = rows.Scan(&result.Email, &result.Amount, &result.Period, &result.Credit,
			&result.CreditQty, &result.Debit, &result.DebitQty, &result.LastUpdated); err != nil {
			return err
		}
		resp = append(resp, *result)
	}

	return r.sendNotification(ctx, accountID, resp)
}

func (e *HtmlElement) String() string {
	return e.string(0)
}

func (e *HtmlElement) string(indent int) string {
	sb := strings.Builder{}
	i := strings.Repeat(" ", indentSize*indent)
	sb.WriteString(fmt.Sprintf("%s<%s>\n",
		i, e.name))
	if len(e.text) > 0 {
		sb.WriteString(strings.Repeat(" ",
			indentSize*(indent+1)))
		sb.WriteString(e.text)
		sb.WriteString("\n")
	}

	for _, el := range e.elements {
		sb.WriteString(el.string(indent + 1))
	}
	sb.WriteString(fmt.Sprintf("%s</%s>\n",
		i, e.name))
	return sb.String()
}

type HtmlBuilder struct {
	rootName string
	root     HtmlElement
}

func NewHtmlBuilder(rootName string) *HtmlBuilder {
	b := HtmlBuilder{rootName,
		HtmlElement{rootName, "", []HtmlElement{}}}
	return &b
}

func (b *HtmlBuilder) String() string {
	return b.root.String()
}

func (b *HtmlBuilder) AddChild(
	childName, childText string) {
	e := HtmlElement{childName, childText, []HtmlElement{}}
	b.root.elements = append(b.root.elements, e)
}

func (b *HtmlBuilder) AddChildFluent(
	childName, childText string) *HtmlBuilder {
	e := HtmlElement{childName, childText, []HtmlElement{}}
	b.root.elements = append(b.root.elements, e)
	return b
}

func htmlBuilder(accountID domain.AccountID, msg []dao) string {

	var (
		totalPeriorAmountCredit, totalPeriorAmountDebit float32
		totalMovementsCredit, totalMovementsDebit       int
	)
	sb := strings.Builder{}
	sb.WriteString(styleHTML)
	sb.WriteString("<h4>")
	sb.WriteString(fmt.Sprintf("balance total up to day: %f usd", msg[0].Amount))
	sb.WriteString("</h4>")
	sb.WriteString("<h5>")
	sb.WriteString(fmt.Sprintf("Account ID: %s ", accountID))
	sb.WriteString("</h5>")

	sb.WriteString("<div class=\"container-tables\">\n    <table class=\"table-balance\">\n        <thead>       \n          <tr>  <th>Period</th>\n            <th>Movements</th>\n            <th>Debit</th>\n            <th>Credit</th>\n            <th>Total Amount</th>\n          </tr>\n        </thead>\n        <tbody>\n            <tr>")
	for _, v := range msg {
		totalPeriorAmountDebit = +v.Debit
		totalPeriorAmountCredit = +v.Credit
		totalMovementsCredit = +v.CreditQty
		totalMovementsDebit = +v.DebitQty

		sb.WriteString("<td>")
		sb.WriteString(v.Period)
		sb.WriteString("</td>")
		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%d", v.CreditQty+v.DebitQty))
		sb.WriteString("</td>")
		sb.WriteString(fmt.Sprintf("%f usd", v.Debit))
		sb.WriteString("</td>")
		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%f usd", v.Credit))
		sb.WriteString("</td>")
	}
	sb.WriteString("  </tr>\n        </tbody>\n    </table>\n</div>\n    <div class=\"total-amount\">")

	sb.WriteString("<p>")
	sb.WriteString(fmt.Sprintf("Average Debit: %f usd", totalPeriorAmountDebit/float32(totalMovementsDebit)))
	sb.WriteString("</p>")

	sb.WriteString("<p>")
	sb.WriteString(fmt.Sprintf("Average Credit: %f usd", totalPeriorAmountCredit/float32(totalMovementsCredit)))
	sb.WriteString("</p>")
	sb.WriteString("</p>")
	sb.WriteString(time.Now().String())
	sb.WriteString("</p>")

	sb.WriteString("    </div>\n    <p>disclaimer: This summary aims to present a fair and unbiased overview, acknowledging both the positive and negative aspects of the discussed topic. While efforts have been made to ensure balance, complexities might lead to nuances being overlooked. Readers are encouraged to conduct further research for a comprehensive understanding. Use this information responsibly.</p>\n</body>\n</html>")

	log.Println(sb.String())
	return sb.String()
}
