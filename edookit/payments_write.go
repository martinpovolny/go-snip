//go:build mutating

// Mutating payment endpoints. Excluded from the default build — the library
// is read-only unless a consumer explicitly opts in with `-tags mutating`.
package edookit

type PrescriptionPerson struct {
	PersonID    int     `json:"person_id"`
	CurrencyID  *int    `json:"currency_id,omitempty"`
	Amount      *string `json:"amount,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CreatePrescriptionReq struct {
	Name            string               `json:"name"`
	Description     *string              `json:"description,omitempty"`
	DueDate         *string              `json:"due_date,omitempty"`
	Amount          *float64             `json:"amount,omitempty"`
	CurrencyID      *int                 `json:"currency_id,omitempty"`
	State           *int                 `json:"state,omitempty"`
	SpecificSymbol  *string              `json:"specific_symbol,omitempty"`
	IsCredit        bool                 `json:"is_credit"`
	PreviousPrescID *int                 `json:"previous_prescription_id,omitempty"`
	Direction       *int                 `json:"direction,omitempty"`
	PersonList      []PrescriptionPerson `json:"person_list,omitempty"`
}

type CreatePaymentReq struct {
	Date                *string `json:"date,omitempty"`
	Amount              float64 `json:"amount"`
	CurrencyID          int     `json:"currency_id"`
	Direction           *int    `json:"direction,omitempty"`
	Type                int     `json:"type"`
	Description         *string `json:"description,omitempty"`
	OrganizationID      *int    `json:"organization_id,omitempty"`
	VariableSymbol      *string `json:"variable_symbol,omitempty"`
	SpecificSymbol      *string `json:"specific_symbol,omitempty"`
	Message             *string `json:"message,omitempty"`
	BankAccountNumber   *string `json:"bank_account_number,omitempty"`
	OperationIdentifier *string `json:"operation_identifier,omitempty"`
	PersonID            *int    `json:"person_id,omitempty"`
	PrescriptionID      *int    `json:"prescription_id,omitempty"`
}

func (c *Client) CreatePaymentPrescription(req CreatePrescriptionReq) (int, error) {
	var resp struct {
		ID int `json:"id"`
	}
	if err := c.post("/api/payment/v1/create/paymentPrescription", req, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (c *Client) CreatePayment(req CreatePaymentReq) (int, error) {
	var resp struct {
		ID int `json:"id"`
	}
	if err := c.post("/api/payment/v1/create/payment", req, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}
