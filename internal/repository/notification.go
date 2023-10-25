package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/castiglionimax/process-csv/internal/domain"
	pkgError "github.com/castiglionimax/process-csv/pkg/error"
	"github.com/harranali/mailing"
	"strings"
	"time"
)

const (
	from = "sender@domain-poc.com"

	getSummary = "SELECT email,amount,period,credit,credit_qty,debit,debit_qty, summaries.last_updated FROM summaries  LEFT JOIN accounts ON summaries.account_id = accounts.id"
)

type (
	dao struct {
		Credit, Debit, Amount float32
		CreditQty, DebitQty   int
		Period, Email         string
		LastUpdated           time.Time
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

	r.mailer.SetSubject("Balance Summary")
	r.mailer.SetHTMLBody(htmlBuilder(accountID, msg))
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
	if len(resp) == 0 {
		return pkgError.HandlerError{Cause: errors.New("not found")}
	}

	return r.sendNotification(ctx, accountID, resp)
}

func htmlBuilder(accountID domain.AccountID, msg []dao) string {

	var (
		totalPeriodAmountCredit, totalPeriodAmountDebit float32
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

	sb.WriteString("<div style=\"width: 100%;\">       <table style=\"font-family: Arial, sans-serif; border-collapse: collapse; width: 100%;\">                    <thead style=\"background-color: #166980; color: #fff; text-align: center;\">\n      \n          <tr>  <th>Period</th>\n            <th>Movements</th>\n            <th>Debit</th>\n            <th>Credit</th>\n            <th>Total Amount</th>\n          </tr>\n        </thead>\n      <tbody style=\"text-align: center;\">")
	for _, v := range msg {
		totalPeriodAmountDebit = +v.Debit
		totalPeriodAmountCredit = +v.Credit
		totalMovementsCredit = +v.CreditQty
		totalMovementsDebit = +v.DebitQty

		sb.WriteString("<tr>")

		sb.WriteString("<td>")
		sb.WriteString(v.Period)
		sb.WriteString("</td>")

		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%d", v.CreditQty+v.DebitQty))
		sb.WriteString("</td>")

		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%f usd", v.Debit))
		sb.WriteString("</td>")

		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%f usd", v.Credit))
		sb.WriteString("</td>")

		sb.WriteString("<td>")
		sb.WriteString(fmt.Sprintf("%f usd", v.Credit+v.Debit))
		sb.WriteString("</td>")

		sb.WriteString("</tr>")
	}
	sb.WriteString("   </tbody>\n    </table>\n</div>\n    <div>")

	sb.WriteString("<div style=\"color: #4b5244; font-size: 15px; font-weight: 700;\">")

	sb.WriteString("<p>")
	sb.WriteString(fmt.Sprintf("Average Debit: %f usd",
		func() float32 {
			if totalPeriodAmountDebit == 0 || totalMovementsDebit == 0 {
				return 0
			}
			return totalPeriodAmountDebit / float32(totalMovementsDebit)
		}()))
	sb.WriteString("</p>")

	sb.WriteString("<p>")
	sb.WriteString(fmt.Sprintf("Average Credit: %f usd",
		func() float32 {
			if totalPeriodAmountCredit == 0 || totalMovementsCredit == 0 {
				return 0
			}
			return totalPeriodAmountCredit / float32(totalMovementsCredit)
		}()))
	sb.WriteString("</p>")

	sb.WriteString("<p>")
	sb.WriteString(time.Now().Format("2006-January-02"))
	sb.WriteString("</p>")

	sb.WriteString("</div>")

	sb.WriteString("    </div>\n    <p>disclaimer: This summary aims to present a fair and unbiased overview, acknowledging both the positive and negative aspects of the discussed topic. While efforts have been made to ensure balance, complexities might lead to nuances being overlooked. Readers are encouraged to conduct further research for a comprehensive understanding. Use this information responsibly.</p>\n</body>\n</html>")

	return sb.String()
}
