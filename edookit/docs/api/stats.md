# Statistics

**Source name**: Statistické informace  
**Auth**: HTTP Basic  
**Base path**: `/api/stats/v1/`

Returns **pseudonymised** (anonymised) statistical data for ministry/reporting purposes.

## Endpoint

### List Student Statistics

```
GET /api/stats/v1/students[/{date}]
```

#### Parameters

| Name | Type | Description |
|------|------|-------------|
| `{date}` | string (path) | Reference date `YYYY-MM-DD`. Defaults to today. |

#### Response Fields

Returns `{"Students": [...]}`. Each record represents one student across the full school history.

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `PersonAnonIdent` | string (UUID) | no | Anonymous person identifier |
| `Age` | int | yes | Age at reference date |
| `Gender` | string | yes | `"M"`, `"F"`, null |
| `ClassName` | string | yes | Class at reference date |
| `EnrolledSince` | string | yes | `YYYY-MM-DD` |
| `InitialGradeNum` | int | yes | Grade when first enrolled |
| `CurrentGradeNum` | int | yes | Current grade |
| `OrganizationName` | string | yes | School name |
| `OrganizationIdent` | string | yes | IZO school identifier |
| `UnenrolledSince` | string | yes | Date of leaving school |
| `UIV_ZPUSOB` | string | yes | Type of compulsory attendance (MŠMT code) |
| `CitizenshipCountryCode` | string | yes | 3-letter ISO |
| `CitizenshipQualifierCode` | string | yes | Citizenship qualifier (MŠMT RAKO code) |
| `PermanentStayCountryCode` | string | yes | 3-letter ISO of permanent residence country |

#### Example

```json
{
  "Students": [
    {
      "PersonAnonIdent": "9dc8eba3-f973-4a1a-af44-9cdb81722f76",
      "Age": 7,
      "Gender": "M",
      "ClassName": "2.B",
      "EnrolledSince": "2018-09-01",
      "CurrentGradeNum": 2,
      "OrganizationName": "ZŠ Edookit",
      "CitizenshipCountryCode": "CZE"
    }
  ]
}
```
