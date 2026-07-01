package edookit

import (
	"fmt"
	"net/url"
)

func (c *Client) ListPaymentPrescriptions(opts PrescriptionListOpts) ([]PaymentPrescription, error) {
	params := url.Values{}
	if opts.DateFrom != nil {
		params.Set("date_from", *opts.DateFrom)
	}
	if opts.DateTo != nil {
		params.Set("date_to", *opts.DateTo)
	}
	if opts.SchoolYear != nil {
		params.Set("school_year", *opts.SchoolYear)
	}
	if opts.PrescriptionID != nil {
		params.Set("prescription_id", fmt.Sprint(*opts.PrescriptionID))
	}
	if opts.PersonID != nil {
		params.Set("person_id", fmt.Sprint(*opts.PersonID))
	}
	if opts.PaymentID != nil {
		params.Set("payment_id", fmt.Sprint(*opts.PaymentID))
	}
	if opts.OperationIdentifier != nil {
		params.Set("operation_identifier", *opts.OperationIdentifier)
	}
	if opts.SpecificSymbol != nil {
		params.Set("specific_symbol", *opts.SpecificSymbol)
	}

	var out []PaymentPrescription
	if err := c.get("/api/payment/v1/list/paymentPrescription", params, &out); err != nil {
		return nil, err
	}
	return out, nil
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

func (c *Client) ListCurrencies() ([]Currency, error) {
	var out []Currency
	if err := c.get("/api/payment/v1/list/currency", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListOrganizations() ([]Organization, error) {
	var out []Organization
	if err := c.get("/api/payment/v1/list/organization", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListPaymentTypes() ([]PaymentType, error) {
	var out []PaymentType
	if err := c.get("/api/payment/v1/list/paymentType", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
