# Payments

**Source name**: Platby  
**Auth**: HTTP Basic  
**Base path**: `/api/payment/v1/`

Supports full CRUD for payment prescriptions and individual payments.

## Read Endpoints

### Get Version
```
GET /api/payment/v1/get/version
```

### List Categories
```
GET /api/payment/v1/list/category
```

### List Currencies
```
GET /api/payment/v1/list/currency
```
Returns `[{"id": 1, "name": "CZK"}, ...]`.

### List Bank Accounts
```
GET /api/payment/v1/list/bankAccount
```
Returns `[{"id": 1, "bankAccountNumber": "123-45678/0100"}, ...]`.

### List Payment Types
```
GET /api/payment/v1/list/paymentType
```
Returns `[{"id": 1, "name": "hotovost", "priority": 400}, ...]`.

### List Organizations
```
GET /api/payment/v1/list/organization
```
Returns `[{"id": 1, "name": "Základní škola Příklad"}, ...]`.

### List Payment Prescriptions
```
GET /api/payment/v1/list/paymentPrescription
```

**JSON body parameters** (all optional, but `date_from`/`date_to` form a range):

| Name | Type | Description |
|------|------|-------------|
| `date_from` | date | Start of due-date range |
| `date_to` | date | End of due-date range |
| `school_year` | string | School year name (cannot combine with date range) |
| `prescription_id` | int | Specific prescription ID |
| `person_id` | int | Filter by person |
| `payment_id` | int | Filter by payment |
| `operation_identifier` | string | Payment operation identifier |
| `specific_symbol` | string | Specific symbol |

---

## Create Endpoints

### Create Payment Prescription
```
POST /api/payment/v1/create/paymentPrescription
```

**JSON body**:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | yes | Prescription name |
| `description` | string | no | Description |
| `due_date` | date | cond. | Due date (must be set either here or per person) |
| `amount` | numeric | cond. | Amount (must be set either here or per person) |
| `currency_id` | int | cond. | Currency ID |
| `state` | int | no | `1` = open, `2` = closed. Default: `1` |
| `specific_symbol` | string | no | Specific symbol |
| `is_credit` | bool | no | Credit prescription? |
| `direction` | int | no | `1` = incoming, `2` = outgoing. Default: `1` |
| `person_list` | array | no | Persons with optional per-person overrides |

**Person list item fields**: `person_id` (required), `currency_id`, `amount`, `due_date`, `description`

### Create Payment
```
POST /api/payment/v1/create/payment
```

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `amount` | numeric | yes | Amount |
| `currency_id` | int | yes | Currency ID |
| `type` | int | yes | Payment type ID |
| `organization_id` | int | no | Organization ID |
| `person_id` | int | no | Person who made the payment |
| `prescription_id` | int | no | Related prescription |
| `date` | date | no | Payment date. Defaults to today. |
| `description` | string | no | Description |
| `variable_symbol` | string | no | Variable symbol |
| `specific_symbol` | string | no | Specific symbol |
| `operation_identifier` | string | no | Operation identifier |

Returns `{"id": <new_payment_id>}`.

---

## Update Endpoints

### Update Payment Prescription
```
POST /api/payment/v1/update/paymentPrescription
```
JSON body: `id` (required) + any of the create fields.

### Update Payment
```
POST /api/payment/v1/update/payment
```
JSON body: `id` (required) + any updatable fields.

---

## Delete Endpoints

### Delete Payment Prescription
```
POST /api/payment/v1/delete/paymentPrescription
```

| Name | Type | Description |
|------|------|-------------|
| `id` | int | Prescription ID |
| `person_list` | int[] | Remove only these persons from prescription |
| `delete_prescription` | bool | If `true`, delete entire prescription. Default: `false` |

### Delete Payment
```
POST /api/payment/v1/delete/payment
```
JSON body: `{"id": <payment_id>}`.
